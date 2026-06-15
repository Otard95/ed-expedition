package database

type StarClass uint8

const (
	StarClassUnknown StarClass = 0x00

	// Main Sequence
	StarClassO StarClass = 0x01
	StarClassB StarClass = 0x02
	StarClassA StarClass = 0x03
	StarClassF StarClass = 0x04
	StarClassG StarClass = 0x05
	StarClassK StarClass = 0x06
	StarClassM StarClass = 0x07

	// Giants & Supergiants
	StarClassKGiant      StarClass = 0x10
	StarClassMGiant      StarClass = 0x11
	StarClassMSuperGiant StarClass = 0x12
	StarClassASuperGiant StarClass = 0x13
	StarClassBSuperGiant StarClass = 0x14
	StarClassFSuperGiant StarClass = 0x15
	StarClassGSuperGiant StarClass = 0x16

	// Brown Dwarfs
	StarClassL StarClass = 0x20
	StarClassT StarClass = 0x21
	StarClassY StarClass = 0x22

	// Carbon Stars
	StarClassC      StarClass = 0x30
	StarClassCN     StarClass = 0x31
	StarClassCJ     StarClass = 0x32
	StarClassMSType StarClass = 0x33
	StarClassSType  StarClass = 0x34

	// White Dwarfs
	StarClassWhiteDwarfD   StarClass = 0x40
	StarClassWhiteDwarfDA  StarClass = 0x41
	StarClassWhiteDwarfDAB StarClass = 0x42
	StarClassWhiteDwarfDAV StarClass = 0x43
	StarClassWhiteDwarfDAZ StarClass = 0x44
	StarClassWhiteDwarfDB  StarClass = 0x45
	StarClassWhiteDwarfDBV StarClass = 0x46
	StarClassWhiteDwarfDBZ StarClass = 0x47
	StarClassWhiteDwarfDC  StarClass = 0x48
	StarClassWhiteDwarfDCV StarClass = 0x49
	StarClassWhiteDwarfDQ  StarClass = 0x4A

	// Wolf-Rayet
	StarClassWolfRayet   StarClass = 0x60
	StarClassWolfRayetC  StarClass = 0x61
	StarClassWolfRayetN  StarClass = 0x62
	StarClassWolfRayetNC StarClass = 0x63
	StarClassWolfRayetO  StarClass = 0x64

	// Proto Stars
	StarClassTTauri   StarClass = 0x70
	StarClassHerbigAe StarClass = 0x71

	// Compact Objects
	StarClassNeutron             StarClass = 0x80
	StarClassBlackHole           StarClass = 0x81
	StarClassSupermassiveBlkHole StarClass = 0x82
)

var starClassMap = map[string]StarClass{
	// Main Sequence
	"O (Blue-White) Star":    StarClassO,
	"B (Blue-White) Star":    StarClassB,
	"A (Blue-White) Star":    StarClassA,
	"F (White) Star":         StarClassF,
	"G (White-Yellow) Star":  StarClassG,
	"K (Yellow-Orange) Star": StarClassK,
	"M (Red dwarf) Star":     StarClassM,

	// Giants & Supergiants
	"K (Yellow-Orange giant) Star":      StarClassKGiant,
	"M (Red giant) Star":                StarClassMGiant,
	"M (Red super giant) Star":          StarClassMSuperGiant,
	"A (Blue-White super giant) Star":   StarClassASuperGiant,
	"B (Blue-White super giant) Star":   StarClassBSuperGiant,
	"F (White super giant) Star":        StarClassFSuperGiant,
	"G (White-Yellow super giant) Star": StarClassGSuperGiant,

	// Brown Dwarfs
	"L (Brown dwarf) Star": StarClassL,
	"T (Brown dwarf) Star": StarClassT,
	"Y (Brown dwarf) Star": StarClassY,

	// Carbon Stars
	"C Star":       StarClassC,
	"CN Star":      StarClassCN,
	"CJ Star":      StarClassCJ,
	"MS-type Star": StarClassMSType,
	"S-type Star":  StarClassSType,

	// White Dwarfs
	"White Dwarf (D) Star":   StarClassWhiteDwarfD,
	"White Dwarf (DA) Star":  StarClassWhiteDwarfDA,
	"White Dwarf (DAB) Star": StarClassWhiteDwarfDAB,
	"White Dwarf (DAV) Star": StarClassWhiteDwarfDAV,
	"White Dwarf (DAZ) Star": StarClassWhiteDwarfDAZ,
	"White Dwarf (DB) Star":  StarClassWhiteDwarfDB,
	"White Dwarf (DBV) Star": StarClassWhiteDwarfDBV,
	"White Dwarf (DBZ) Star": StarClassWhiteDwarfDBZ,
	"White Dwarf (DC) Star":  StarClassWhiteDwarfDC,
	"White Dwarf (DCV) Star": StarClassWhiteDwarfDCV,
	"White Dwarf (DQ) Star":  StarClassWhiteDwarfDQ,

	// Wolf-Rayet
	"Wolf-Rayet Star":    StarClassWolfRayet,
	"Wolf-Rayet C Star":  StarClassWolfRayetC,
	"Wolf-Rayet N Star":  StarClassWolfRayetN,
	"Wolf-Rayet NC Star": StarClassWolfRayetNC,
	"Wolf-Rayet O Star":  StarClassWolfRayetO,

	// Proto Stars
	"T Tauri Star":      StarClassTTauri,
	"Herbig Ae/Be Star": StarClassHerbigAe,

	// Compact Objects
	"Neutron Star":            StarClassNeutron,
	"Black Hole":              StarClassBlackHole,
	"Supermassive Black Hole": StarClassSupermassiveBlkHole,
}

var starClassNameMap = reverseStarClassMap(starClassMap)

func reverseStarClassMap(src map[string]StarClass) map[StarClass]string {
	dst := make(map[StarClass]string, len(src))
	for name, class := range src {
		if _, exists := dst[class]; exists {
			panic("duplicate star class code")
		}
		dst[class] = name
	}
	return dst
}

func parseStarClass(starType string) StarClass {
	if class, ok := starClassMap[starType]; ok {
		return class
	}
	return StarClassUnknown
}

func StarClassName(class StarClass) string {
	return starClassNameMap[class]
}

func IsScoopableStarClass(class StarClass) bool {
	return (class >= StarClassO && class <= StarClassM) || (class >= StarClassKGiant && class <= StarClassGSuperGiant)
}
