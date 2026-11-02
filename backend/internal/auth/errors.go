// Package auth concentra la logica de autenticacion del backend:
// hashing de passwords, emision y validacion de JWT, y errores tipados
// para que los services y middlewares decidan que respuesta dar al cliente.
package auth

import "errors"

// ErrInvalidCredentials se devuelve cuando el email o el password no coinciden.
// El service que recibe este error debe mapearlo a "invalid_credentials"
// SIN exponer el detalle (no diferenciar entre "email no existe" y
// "password incorrecto", para no facilitar enumeracion de cuentas).
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// ErrInvalidToken se devuelve cuando un JWT viene mal firmado, vencido o malformado.
var ErrInvalidToken = errors.New("auth: invalid token")
