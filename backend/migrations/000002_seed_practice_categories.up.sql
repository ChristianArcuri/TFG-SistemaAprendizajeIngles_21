-- Categorias del MVP. Idempotente: si la migracion se reaplica
-- (por ejemplo en tests), no duplica datos.
-- Alcance del prototipo TFG: entrevista RRHH, daily meeting y quiz de ingles IT.
INSERT INTO practice_categories (code, name, description, icon) VALUES
    ('hr_interview',  'Entrevista con Recursos Humanos', 'Practica comprender la pregunta y responder en ingles en una entrevista laboral.', 'briefcase'),
    ('daily_meeting', 'Daily Meeting',                   'Practica explicar avances, proximas tareas y bloqueos en una reunion diaria.',     'calendar'),
    ('it_quiz',       'Quiz de Ingles IT',               'Elegi la expresion correcta de vocabulario tecnico dentro de una situacion IT.',   'check-circle')
ON CONFLICT (code) DO NOTHING;
