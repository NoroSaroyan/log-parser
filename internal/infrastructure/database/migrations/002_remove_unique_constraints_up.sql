-- Remove incorrect unique constraints that prevent legitimate duplicate PCBA numbers
-- PCBA numbers can appear multiple times due to retests, quality checks, and different test stations

-- Drop unique constraint on logistic_data.pcba_number
-- This allows the same PCBA to be tested multiple times
ALTER TABLE logistic_data DROP CONSTRAINT IF EXISTS logistic_data_pcba_number_key;

-- Drop unique constraint on download_info.tcu_pcba_number  
-- This allows multiple download records for the same PCBA number
ALTER TABLE download_info DROP CONSTRAINT IF EXISTS download_info_tcu_pcba_number_key;