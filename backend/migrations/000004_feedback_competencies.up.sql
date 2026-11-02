-- La base creada con el esquema anterior guarda grammar/fluency/pronunciation.
-- El prototipo usa comprension, vocabulario profesional y comunicacion profesional.
-- En una base nueva estas columnas ya existen y este paso no cambia nada.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'ai_feedback_results'
          AND column_name = 'grammar_score'
    ) THEN
        ALTER TABLE ai_feedback_results
            ADD COLUMN IF NOT EXISTS comprehension_score INT,
            ADD COLUMN IF NOT EXISTS communication_score INT;

        UPDATE ai_feedback_results
        SET comprehension_score = grammar_score
        WHERE comprehension_score IS NULL;

        UPDATE ai_feedback_results
        SET communication_score = fluency_score
        WHERE communication_score IS NULL;

        ALTER TABLE ai_feedback_results
            DROP COLUMN grammar_score,
            DROP COLUMN fluency_score,
            DROP COLUMN pronunciation_score;
    END IF;
END $$;
