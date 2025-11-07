-- Initial database schema
-- Creates all base tables for the log parser application

CREATE TABLE download_info
(
    id                     SERIAL PRIMARY KEY,
    test_station           TEXT NOT NULL,
    flash_entity_type      TEXT,
    tcu_pcba_number        TEXT NOT NULL UNIQUE,
    flash_elapsed_time     INTEGER,
    tcu_entity_flash_state TEXT,
    part_number            TEXT,
    product_line           TEXT,
    download_tool_version  TEXT,
    download_finished_time TEXT
);

CREATE TABLE logistic_data
(
    id                            SERIAL PRIMARY KEY,
    pcba_number                   TEXT NOT NULL UNIQUE,
    product_sn                    TEXT,
    part_number                   TEXT,
    vp_app_version                TEXT,
    vp_boot_loader_version        TEXT,
    vp_core_version               TEXT,
    supplier_hardware_version     TEXT,
    manufacturer_hardware_version TEXT,
    manufacturer_software_version TEXT,
    ble_mac                       TEXT,
    ble_sn                        TEXT,
    ble_version                   TEXT,
    ble_passwork_key              TEXT,
    ap_app_version                TEXT,
    ap_kernel_version             TEXT,
    tcu_iccid                     TEXT,
    phone_number                  TEXT,
    imei                          TEXT,
    imsi                          TEXT,
    production_date               TEXT
);

CREATE TABLE test_station_record
(
    id                 SERIAL PRIMARY KEY,
    part_number        TEXT,
    test_station       TEXT    NOT NULL,
    entity_type        TEXT,
    product_line       TEXT,
    test_tool_version  TEXT,
    test_finished_time TEXT,
    is_all_passed      BOOLEAN,
    error_codes        TEXT,
    logistic_data_id   INTEGER NOT NULL,
    CONSTRAINT fk_logistic_data
        FOREIGN KEY (logistic_data_id)
            REFERENCES logistic_data (id)
            ON DELETE CASCADE
);

CREATE TABLE test_step
(
    id                     SERIAL PRIMARY KEY,
    test_step_name         TEXT    NOT NULL,
    test_threshold_value   TEXT,
    test_measured_value    TEXT,
    test_step_elapsed_time INTEGER,
    test_step_result       TEXT,
    test_step_error_code   TEXT,
    test_station_record_id INTEGER NOT NULL,
    CONSTRAINT fk_test_station_record
        FOREIGN KEY (test_station_record_id)
            REFERENCES test_station_record (id)
            ON DELETE CASCADE
);