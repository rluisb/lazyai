package sqlite

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
