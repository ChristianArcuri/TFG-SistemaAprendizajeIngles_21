package users

import "errors"

// ErrEmailAlreadyExists indica que se intento registrar un email ya tomado.
var ErrEmailAlreadyExists = errors.New("users: email already exists")

// ErrUserNotFound indica que el repositorio no encontro el usuario buscado.
var ErrUserNotFound = errors.New("users: user not found")

// ErrProfileNotFound indica que el usuario no tiene perfil cargado.
var ErrProfileNotFound = errors.New("users: profile not found")
