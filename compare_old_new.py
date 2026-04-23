#!/usr/bin/env python3
"""
Comprehensive diff across old and new log files.

For each file, compute:
  - per-type payload counts (Download / PCBA-station / Final-station / PCBA-steps / Final-steps)
  - PCBA with missing StationInformation record (Bug #1): split by missing PCBA vs Final
  - PCBA with retry-asymmetry inside same type (Bug #2)
  - PCBA with station records that have empty PCBANumber (Problem #3)
  - TestToolVersion distribution (PCBA and Final separately)
  - Unique TestStepName count + set
  - Unique TestStation values
  - Total number of devices with full PCBA+Final coverage

Output: a side-by-side comparison printed as plain text.

Handles both .gz and plain-text files.
"""
import gzip
import re
import glob
import os
from collections import defaultdict


OLD_DIR = "corporate_resources/old_logs"
NEW_DIR = "corporate_resources"

SCAN_TO_TYPE = {
    "PCBA Scan": "PCBA",
    "Compare PCBA Serial Number": "Final",
    "Valid PCBA Serial Number": "Final",
}


def open_log(path):
    if path.endswith(".gz"):
        return gzip.open(path, "rt", errors="replace")
    return open(path, "rt", errors="replace")


def extract_blocks_iter(path):
    """Iterate 'Data { ... }' and 'Data [ ... ]' blocks from the log file."""
    with open_log(path) as fp:
        lines = fp.readlines()

    inside_obj = False
    inside_arr = False
    brace = 0
    bracket = 0
    block = []

    for raw in lines:
        if " Data  {" in raw and not inside_obj and not inside_arr:
            inside_obj = True
            brace = 0
            block = []
            idx = raw.index(" Data  ")
            part = raw[idx + 7:]
            block.append(part)
            brace += part.count("{") - part.count("}")
            if brace <= 0 and len("".join(block).strip()) > 10:
                yield "obj", "".join(block)
                inside_obj = False; block = []
        elif " Data  [" in raw and not inside_obj and not inside_arr:
            inside_arr = True
            bracket = 0
            block = []
            idx = raw.index(" Data  ")
            part = raw[idx + 7:]
            block.append(part)
            bracket += part.count("[") - part.count("]")
            if bracket <= 0 and len("".join(block).strip()) > 10:
                yield "arr", "".join(block)
                inside_arr = False; block = []
        elif inside_obj:
            p = raw.find("]:")
            if p != -1 and len(raw) > p + 2:
                part = raw[p + 2:]
                block.append(part)
                brace += part.count("{") - part.count("}")
            if brace <= 0 and len("".join(block).strip()) > 10:
                yield "obj", "".join(block)
                inside_obj = False; block = []
        elif inside_arr:
            p = raw.find("]:")
            if p != -1 and len(raw) > p + 2:
                part = raw[p + 2:]
                block.append(part)
                bracket += part.count("[") - part.count("]")
            if bracket <= 0 and len("".join(block).strip()) > 10:
                yield "arr", "".join(block)
                inside_arr = False; block = []


def analyze(path):
    per_pcba_station = defaultdict(lambda: defaultdict(int))  # pcba -> type -> count
    per_pcba_steps = defaultdict(lambda: defaultdict(int))    # pcba -> type -> count
    downloads_for_pcba = defaultdict(int)
    stations_empty_pcba = 0
    testtool_versions = defaultdict(lambda: defaultdict(int)) # station_type -> version -> count
    step_names = set()
    station_values = set()
    type_counts = {"Download": 0, "PCBA_station": 0, "Final_station": 0, "PCBA_steps": 0, "Final_steps": 0, "other_station": 0, "unknown_steps": 0}
    noscan_steps = 0

    for kind, text in extract_blocks_iter(path):
        if kind == "obj":
            m_ts = re.search(r'"TestStation"\s*:\s*"([^"]+)"', text)
            if not m_ts:
                continue
            ts = m_ts.group(1).strip()
            station_values.add(ts)
            if ts == "Download":
                type_counts["Download"] += 1
                m = re.search(r'"TcuPCBANumber"\s*:\s*"([^"]+)"', text)
                if m:
                    downloads_for_pcba[m.group(1).strip()] += 1
                continue
            if ts in ("PCBA", "Final"):
                m_pcba = re.search(r'"PCBANumber"\s*:\s*"([^"]*)"', text)
                pcba = m_pcba.group(1).strip() if m_pcba else ""
                if not pcba:
                    stations_empty_pcba += 1
                    # skip indexing — these pollute logistic_data but don't
                    # participate in Bug #1 matching.
                    type_counts[f"{ts}_station"] += 1
                    continue
                m_ttv = re.search(r'"TestToolVersion"\s*:\s*"([^"]*)"', text)
                ttv = m_ttv.group(1).strip() if m_ttv else "<none>"
                testtool_versions[ts][ttv] += 1
                per_pcba_station[pcba][ts] += 1
                type_counts[f"{ts}_station"] += 1
            else:
                type_counts["other_station"] += 1

        elif kind == "arr":
            stype = None
            pcba = ""
            for m in re.finditer(r'"TestStepName"\s*:\s*"([^"]+)"', text):
                name = m.group(1)
                step_names.add(name)
                if stype is None and name in SCAN_TO_TYPE:
                    stype = SCAN_TO_TYPE[name]
                    after = text[m.end():]
                    mv = re.search(r'"TestMeasuredValue"\s*:\s*"?([^",\s}]+)"?', after)
                    if mv:
                        pcba = mv.group(1).strip()
            if stype and pcba:
                per_pcba_steps[pcba][stype] += 1
                type_counts[f"{stype}_steps"] += 1
            else:
                type_counts["unknown_steps"] += 1
                noscan_steps += 1

    # Compute per-PCBA issues
    all_pcbas = set(per_pcba_station.keys()) | set(per_pcba_steps.keys())
    bug1_missing_pcba_record = []
    bug1_missing_final_record = []
    bug1_total_orphan = []          # steps with zero records of any type
    bug2_retry_asymmetry = defaultdict(int)  # stype -> count
    full_coverage = 0

    for pcba in all_pcbas:
        st = per_pcba_station[pcba]
        stp = per_pcba_steps[pcba]
        has_p_st = st.get("PCBA", 0) > 0
        has_f_st = st.get("Final", 0) > 0
        has_p_stp = stp.get("PCBA", 0) > 0
        has_f_stp = stp.get("Final", 0) > 0

        if has_p_stp and not has_p_st:
            if has_f_st:
                bug1_missing_pcba_record.append(pcba)
            else:
                bug1_total_orphan.append(pcba)
        if has_f_stp and not has_f_st:
            if has_p_st:
                bug1_missing_final_record.append(pcba)
            else:
                # already counted in bug1_total_orphan if neither type exists
                if pcba not in bug1_total_orphan:
                    bug1_total_orphan.append(pcba)

        # Bug #2: same-type asymmetry
        for t in ("PCBA", "Final"):
            if st.get(t, 0) > 0 and stp.get(t, 0) > 0 and st[t] != stp[t]:
                bug2_retry_asymmetry[t] += 1

        if has_p_st and has_f_st and has_p_stp and has_f_stp:
            full_coverage += 1

    return {
        "path": path,
        "type_counts": type_counts,
        "stations_empty_pcba": stations_empty_pcba,
        "unique_pcbas": len(all_pcbas),
        "full_coverage": full_coverage,
        "bug1_missing_pcba_record": len(bug1_missing_pcba_record),
        "bug1_missing_final_record": len(bug1_missing_final_record),
        "bug1_total_orphan": len(bug1_total_orphan),
        "bug2_retry_asymmetry_pcba": bug2_retry_asymmetry.get("PCBA", 0),
        "bug2_retry_asymmetry_final": bug2_retry_asymmetry.get("Final", 0),
        "testtool_versions": {k: dict(v) for k, v in testtool_versions.items()},
        "step_names": step_names,
        "station_values": station_values,
        "noscan_steps": noscan_steps,
        "bug1_sample_missing_pcba": bug1_missing_pcba_record[:5],
        "bug1_sample_missing_final": bug1_missing_final_record[:5],
        "bug1_sample_orphan": bug1_total_orphan[:5],
    }


def fmt_header(label):
    return f"\n{'='*8} {label} {'='*8}"


def main():
    old_files = sorted(glob.glob(f"{OLD_DIR}/*"))
    new_files = sorted(glob.glob(f"{NEW_DIR}/mesrestapi.log-*.gz"))
    new_files = [f for f in new_files if "old_logs" not in f]

    results = []

    print(fmt_header("OLD LOGS"))
    for path in old_files:
        print(f"[scanning] {os.path.basename(path)}...", flush=True)
        try:
            r = analyze(path)
            r["era"] = "old"
            results.append(r)
        except Exception as e:
            print(f"  ERROR: {e}")

    print(fmt_header("NEW LOGS"))
    for path in new_files:
        print(f"[scanning] {os.path.basename(path)}...", flush=True)
        try:
            if os.path.getsize(path) == 0:
                continue
            r = analyze(path)
            r["era"] = "new"
            results.append(r)
        except Exception as e:
            print(f"  ERROR: {e}")

    # ===== Side-by-side table =====
    print(fmt_header("Per-file summary"))
    cols = [
        ("file",                         lambda r: os.path.basename(r["path"])),
        ("download",                     lambda r: r["type_counts"]["Download"]),
        ("PCBA-st",                      lambda r: r["type_counts"]["PCBA_station"]),
        ("Final-st",                     lambda r: r["type_counts"]["Final_station"]),
        ("PCBA-steps",                   lambda r: r["type_counts"]["PCBA_steps"]),
        ("Final-steps",                  lambda r: r["type_counts"]["Final_steps"]),
        ("emptyPCBA",                    lambda r: r["stations_empty_pcba"]),
        ("full-cov",                     lambda r: r["full_coverage"]),
        ("bug1:missPCBA",                lambda r: r["bug1_missing_pcba_record"]),
        ("bug1:missFinal",               lambda r: r["bug1_missing_final_record"]),
        ("bug1:orphan",                  lambda r: r["bug1_total_orphan"]),
        ("bug2:asymPCBA",                lambda r: r["bug2_retry_asymmetry_pcba"]),
        ("bug2:asymFinal",               lambda r: r["bug2_retry_asymmetry_final"]),
    ]
    # Print header
    print(" | ".join(f"{c[0]:>15}" if i > 0 else f"{c[0]:>36}" for i, c in enumerate(cols)))
    print("-" * 250)
    for r in results:
        print(" | ".join(f"{c[1](r):>15}" if i > 0 else f"{c[1](r):>36}" for i, c in enumerate(cols)))

    # ===== TestToolVersion distribution =====
    print(fmt_header("TestToolVersion distribution"))
    for r in results:
        tts = r["testtool_versions"]
        print(f"  {os.path.basename(r['path']):40s}  "
              f"PCBA: {dict(tts.get('PCBA', {}))}  "
              f"Final: {dict(tts.get('Final', {}))}")

    # ===== TestStation values =====
    print(fmt_header("TestStation values (union per file)"))
    for r in results:
        print(f"  {os.path.basename(r['path']):40s}  {sorted(r['station_values'])}")

    # ===== Unique TestStepName counts + cross-file diff =====
    print(fmt_header("Unique TestStepName counts"))
    for r in results:
        print(f"  {os.path.basename(r['path']):40s}  {len(r['step_names'])} unique")

    # Union of all old, all new
    old_union = set()
    new_union = set()
    for r in results:
        if r["era"] == "old":
            old_union |= r["step_names"]
        else:
            new_union |= r["step_names"]
    print(f"\n  OLD union: {len(old_union)} unique")
    print(f"  NEW union: {len(new_union)} unique")
    only_old = old_union - new_union
    only_new = new_union - old_union
    print(f"  in OLD but not in NEW ({len(only_old)}): {sorted(only_old)}")
    print(f"  in NEW but not in OLD ({len(only_new)}): {sorted(only_new)}")

    # ===== Bug samples =====
    print(fmt_header("Bug sample PCBAs (first ~5 per file)"))
    for r in results:
        print(f"  {os.path.basename(r['path'])}:")
        if r['bug1_sample_missing_pcba']: print(f"    missing PCBA record: {r['bug1_sample_missing_pcba']}")
        if r['bug1_sample_missing_final']: print(f"    missing Final record: {r['bug1_sample_missing_final']}")
        if r['bug1_sample_orphan']: print(f"    total orphan (no station at all): {r['bug1_sample_orphan']}")


if __name__ == "__main__":
    main()
