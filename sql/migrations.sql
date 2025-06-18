-- Migration to add due_date column
ALTER TABLE todos ADD COLUMN due_date DATETIME;

-- Migration to add scheduled_start column  
ALTER TABLE todos ADD COLUMN scheduled_start DATETIME;

-- Migration to add scheduled_end column
ALTER TABLE todos ADD COLUMN scheduled_end DATETIME;
