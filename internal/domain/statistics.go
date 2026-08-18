package domain

// Statistics representa los datos devueltos por el servicio de estadísticas
// (API Node.js). Se modela como un mapa para no acoplarse a la estructura
// exacta que produce el servicio remoto.
type Statistics map[string]interface{}
