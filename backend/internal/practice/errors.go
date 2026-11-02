package practice

import "errors"

// ErrCategoryNotFound se devuelve cuando un category code no matchea.
var ErrCategoryNotFound = errors.New("practice: category not found")

// ErrPromptNotFound se devuelve cuando se intenta usar un prompt que no existe.
var ErrPromptNotFound = errors.New("practice: prompt not found")

// ErrPromptCategoryMismatch se devuelve cuando el prompt indicado existe
// pero pertenece a otra categoria que la enviada en CreateSession.
var ErrPromptCategoryMismatch = errors.New("practice: prompt does not belong to category")

// ErrSessionNotFound se devuelve cuando una sesion no existe o no pertenece
// al usuario que la pide. Consolidamos los dos casos para no facilitar
// enumeracion de IDs ajenos.
var ErrSessionNotFound = errors.New("practice: session not found")

// ErrFeedbackNotFound se devuelve cuando una sesion todavia no tiene feedback.
var ErrFeedbackNotFound = errors.New("practice: feedback not found")

// ErrAudioTooLarge se devuelve cuando el audio supera MAX_AUDIO_BYTES.
var ErrAudioTooLarge = errors.New("practice: audio exceeds maximum size")

// ErrUnsupportedAudioType se devuelve cuando el MIME del audio no esta permitido.
var ErrUnsupportedAudioType = errors.New("practice: unsupported audio mime type")

// ErrEmptyUserText se devuelve cuando userText llega vacio en /analyze-text.
var ErrEmptyUserText = errors.New("practice: user text is empty")

// ErrUserTextTooLong se devuelve cuando userText supera el limite permitido.
var ErrUserTextTooLong = errors.New("practice: user text exceeds maximum length")

// ErrInvalidQuizResult se devuelve cuando el resumen del quiz no es coherente.
var ErrInvalidQuizResult = errors.New("practice: invalid quiz result")
