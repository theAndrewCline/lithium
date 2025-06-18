UPDATE todos 
SET done = NOT done, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?
