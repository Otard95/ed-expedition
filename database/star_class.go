package database

var starClassMap = map[string]uint8{
	// Main Sequence
	"O (Blue-White) Star":    0x01,
	"B (Blue-White) Star":    0x02,
	"A (Blue-White) Star":    0x03,
	"F (White) Star":         0x04,
	"G (White-Yellow) Star":  0x05,
	"K (Yellow-Orange) Star": 0x06,
	"M (Red dwarf) Star":     0x07,

	// Giants & Supergiants
	"K (Yellow-Orange giant) Star":      0x10,
	"M (Red giant) Star":                0x11,
	"M (Red super giant) Star":          0x12,
	"A (Blue-White super giant) Star":   0x13,
	"B (Blue-White super giant) Star":   0x14,
	"F (White super giant) Star":        0x15,
	"G (White-Yellow super giant) Star": 0x16,

	// Brown Dwarfs
	"L (Brown dwarf) Star": 0x20,
	"T (Brown dwarf) Star": 0x21,
	"Y (Brown dwarf) Star": 0x22,

	// Carbon Stars
	"C Star":       0x30,
	"CN Star":      0x31,
	"CJ Star":      0x32,
	"MS-type Star": 0x33,
	"S-type Star":  0x34,

	// White Dwarfs
	"White Dwarf (D) Star":   0x40,
	"White Dwarf (DA) Star":  0x41,
	"White Dwarf (DAB) Star": 0x42,
	"White Dwarf (DAV) Star": 0x43,
	"White Dwarf (DAZ) Star": 0x44,
	"White Dwarf (DB) Star":  0x45,
	"White Dwarf (DBV) Star": 0x46,
	"White Dwarf (DBZ) Star": 0x47,
	"White Dwarf (DC) Star":  0x48,
	"White Dwarf (DCV) Star": 0x49,
	"White Dwarf (DQ) Star":  0x4A,

	// Wolf-Rayet
	"Wolf-Rayet Star":    0x60,
	"Wolf-Rayet C Star":  0x61,
	"Wolf-Rayet N Star":  0x62,
	"Wolf-Rayet NC Star": 0x63,
	"Wolf-Rayet O Star":  0x64,

	// Proto Stars
	"T Tauri Star":      0x70,
	"Herbig Ae/Be Star": 0x71,

	// Compact Objects
	"Neutron Star":            0x80,
	"Black Hole":              0x81,
	"Supermassive Black Hole": 0x82,
}

var starClassNameMap = reverseStarClassMap(starClassMap)

func reverseStarClassMap(src map[string]uint8) map[uint8]string {
	dst := make(map[uint8]string, len(src))
	for name, class := range src {
		if _, exists := dst[class]; exists {
			panic("duplicate star class code")
		}
		dst[class] = name
	}
	return dst
}

func parseStarClass(starType string) uint8 {
	if starType == "" {
		return 0x00
	}
	if class, ok := starClassMap[starType]; ok {
		return class
	}
	return 0x00
}

func StarClassName(class uint8) string {
	return starClassNameMap[class]
}

func IsScoopableStarClass(class uint8) bool {
	return (class >= 0x01 && class <= 0x07) || (class >= 0x10 && class <= 0x16)
}
