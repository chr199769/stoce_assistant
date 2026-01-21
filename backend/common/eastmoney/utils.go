package eastmoney

import "strings"

func getSecId(code string) string {
	cleanCode := code
	prefix := ""
	
	// Handle sh/sz prefix
	if len(code) > 2 {
		p := strings.ToLower(code[:2])
		if p == "sh" || p == "sz" {
			prefix = p
			cleanCode = code[2:]
		}
	}
	
	// If already has prefix logic, use it
	if prefix == "sh" {
		return "1." + cleanCode
	} else if prefix == "sz" {
		return "0." + cleanCode
	}

	// Heuristic based on first digit
	if strings.HasPrefix(cleanCode, "6") {
		return "1." + cleanCode
	}
	if strings.HasPrefix(cleanCode, "0") || strings.HasPrefix(cleanCode, "3") {
		return "0." + cleanCode
	}
	if strings.HasPrefix(cleanCode, "8") || strings.HasPrefix(cleanCode, "4") {
		return "0." + cleanCode // BJ usually 0 too in Eastmoney? Actually BJ is often 0.
	}
	
	// Default to 0
	return "0." + cleanCode
}

func getCleanCode(code string) string {
	if len(code) > 6 {
		return code[len(code)-6:]
	}
	return code
}
