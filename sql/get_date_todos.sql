SELECT id, title, description, done, due_date, scheduled_start, scheduled_end, created_at, updated_at 
FROM todos 
WHERE scheduled_start IS NOT NULL AND DATE(scheduled_start) = DATE(?) 
ORDER BY scheduled_start ASC
