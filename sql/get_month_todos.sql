SELECT id, title, description, done, due_date, scheduled_start, scheduled_end, created_at, updated_at 
FROM todos 
WHERE scheduled_start IS NOT NULL 
  AND strftime('%Y-%m', scheduled_start) = strftime('%Y-%m', ?)
ORDER BY scheduled_start ASC
