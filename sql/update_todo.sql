UPDATE todos 
SET title = ?, description = ?, due_date = ?, scheduled_start = ?, scheduled_end = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?
