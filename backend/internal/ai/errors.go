package ai

import "errors"

// ErrProviderFailure indica que el proveedor de IA fallo. Los callers (p. ej.
// practice.Service) deben mapearlo a un internal_error sin reenviar el mensaje
// original de OpenAI al cliente.
var ErrProviderFailure = errors.New("ai: provider failure")
