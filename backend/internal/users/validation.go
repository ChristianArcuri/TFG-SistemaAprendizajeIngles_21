package users

import "strings"

// normalizeEmail aplica la regla del proyecto: emails en minusculas y
// sin espacios laterales. La unicidad esta protegida por el indice UNIQUE
// pero queremos que "User@example.com" y "user@example.com" colisionen.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// trim recorta espacios pero conserva el contenido. Util para campos de perfil.
func trim(s string) string {
	return strings.TrimSpace(s)
}
