-- Rollback: Re-add unique constraints on PCBA number fields
-- WARNING: This rollback may fail if duplicate PCBA numbers exist in the data

-- Re-add unique constraint on logistic_data.pcba_number
-- This will fail if duplicate PCBA numbers exist
ALTER TABLE logistic_data ADD CONSTRAINT logistic_data_pcba_number_key UNIQUE (pcba_number);

-- Re-add unique constraint on download_info.tcu_pcba_number
-- This will fail if duplicate tcu_pcba_numbers exist  
ALTER TABLE download_info ADD CONSTRAINT download_info_tcu_pcba_number_key UNIQUE (tcu_pcba_number);