DELETE FROM practice_categories
WHERE code IN (
    'hr_interview',
    'tech_interview',
    'daily_meeting',
    'client_support',
    'writing_email',
    'it_quiz'
);
