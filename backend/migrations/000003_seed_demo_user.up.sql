-- Usuario fijo para la demo del TFG.
-- Email: admintfg@gmail.com
-- Password: AdminTFG
-- El hash es bcrypt (costo 12). No se guarda la contraseña en texto plano.
INSERT INTO users (email, password_hash, full_name)
VALUES (
    'admintfg@gmail.com',
    '$2a$12$.Wh0KAzfGVfd2gaccaOFxeJSejGOpWRgMliM4Jv7OND3ydG5XufWi',
    'Admin TFG'
)
ON CONFLICT (email) DO NOTHING;
