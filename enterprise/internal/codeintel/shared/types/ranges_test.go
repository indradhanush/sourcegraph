package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/scip/bindings/go/scip"
)

func TestRangeEncoding(t *testing.T) {
	ranges := []int32{
		0, 8, 0, 18,
		3, 2, 3, 9,
		4, 2, 4, 5,
		5, 2, 5, 6,
		6, 2, 6, 6,
		8, 2, 8, 32,
		9, 1, 9, 6,
		9, 8, 9, 49,
		11, 2, 11, 28,
		13, 2, 13, 53,
		14, 2, 14, 54,
		15, 2, 15, 57,
		16, 2, 16, 51,
		17, 2, 17, 54,
		18, 2, 18, 58,
		19, 2, 19, 47,
		23, 4, 23, 23,
		23, 26, 23, 32,
		23, 33, 23, 36,
		26, 5, 26, 11,
		26, 12, 26, 13,
		26, 14, 26, 20,
		27, 1, 27, 6,
		27, 18, 27, 23,
		27, 24, 27, 25,
		28, 1, 28, 8,
		28, 18, 28, 25,
		28, 26, 28, 27,
		29, 1, 29, 8,
		29, 18, 29, 31,
		30, 1, 30, 13,
		30, 18, 30, 23,
		30, 24, 30, 29,
		31, 1, 31, 15,
		31, 18, 31, 23,
		31, 24, 31, 29,
		32, 1, 32, 14,
		32, 18, 32, 23,
		32, 24, 32, 29,
		33, 1, 33, 12,
		34, 1, 34, 17,
		35, 1, 35, 8,
		35, 18, 35, 25,
		35, 26, 35, 33,
		36, 1, 36, 11,
		36, 18, 36, 25,
		36, 26, 36, 33,
		37, 1, 37, 14,
		38, 1, 38, 3,
		38, 18, 38, 22,
		38, 23, 38, 32,
		39, 1, 39, 9,
		40, 1, 40, 13,
		40, 19, 40, 24,
		43, 5, 43, 18,
		47, 1, 47, 5,
		52, 1, 52, 15,
		57, 1, 57, 12,
		63, 1, 63, 13,
		69, 1, 69, 14,
		69, 15, 69, 19,
		69, 20, 69, 28,
		72, 1, 72, 9,
		72, 10, 72, 14,
		72, 15, 72, 23,
		77, 1, 77, 18,
		77, 19, 77, 23,
		77, 24, 77, 32,
		80, 1, 80, 21,
		80, 22, 80, 26,
		80, 27, 80, 35,
		83, 1, 83, 8,
		83, 9, 83, 28,
		86, 5, 86, 14,
		86, 15, 86, 16,
		86, 17, 86, 23,
		86, 25, 86, 28,
		86, 29, 86, 36,
		86, 37, 86, 44,
		86, 46, 86, 51,
		86, 52, 86, 57,
		86, 58, 86, 59,
		86, 62, 86, 69,
		86, 70, 86, 77,
		86, 78, 86, 79,
		86, 82, 86, 89,
		86, 90, 86, 103,
		86, 106, 86, 112,
		86, 113, 86, 114,
		87, 1, 87, 6,
		87, 10, 87, 15,
		87, 16, 87, 28,
		88, 8, 88, 17,
		88, 18, 88, 21,
		88, 23, 88, 28,
		88, 30, 88, 37,
		88, 39, 88, 46,
		88, 48, 88, 53,
		88, 55, 88, 60,
		88, 62, 88, 67,
		91, 5, 91, 14,
		91, 15, 91, 16,
		91, 17, 91, 23,
		91, 25, 91, 28,
		91, 29, 91, 36,
		91, 37, 91, 44,
		91, 46, 91, 51,
		91, 52, 91, 57,
		91, 58, 91, 59,
		91, 62, 91, 69,
		91, 70, 91, 77,
		91, 78, 91, 79,
		91, 82, 91, 89,
		91, 90, 91, 103,
		91, 105, 91, 114,
		91, 116, 91, 130,
		91, 132, 91, 145,
		91, 146, 91, 151,
		91, 152, 91, 157,
		91, 160, 91, 166,
		91, 167, 91, 168,
		92, 4, 92, 11,
		92, 12, 92, 16,
		95, 4, 95, 11,
		95, 12, 95, 26,
		96, 2, 96, 9,
		96, 10, 96, 24,
		96, 27, 96, 35,
		96, 36, 96, 39,
		100, 4, 100, 11,
		100, 12, 100, 19,
		100, 20, 100, 26,
		101, 2, 101, 9,
		101, 10, 101, 17,
		101, 18, 101, 24,
		101, 27, 101, 30,
		101, 31, 101, 37,
		101, 48, 101, 55,
		101, 56, 101, 60,
		101, 86, 101, 93,
		101, 94, 101, 108,
		103, 1, 103, 8,
		103, 9, 103, 16,
		103, 17, 103, 23,
		103, 26, 103, 33,
		103, 34, 103, 41,
		103, 42, 103, 48,
		103, 49, 103, 53,
		103, 54, 103, 57,
		103, 58, 103, 64,
		103, 73, 103, 80,
		103, 81, 103, 85,
		105, 1, 105, 15,
		105, 17, 105, 23,
		105, 27, 105, 34,
		105, 35, 105, 45,
		105, 46, 105, 49,
		107, 1, 107, 17,
		107, 41, 107, 48,
		107, 49, 107, 60,
		108, 5, 108, 6,
		108, 13, 108, 14,
		108, 17, 108, 24,
		108, 25, 108, 36,
		108, 38, 108, 39,
		109, 2, 109, 18,
		112, 9, 112, 15,
		112, 16, 112, 17,
		113, 2, 113, 7,
		113, 20, 113, 25,
		114, 2, 114, 9,
		114, 20, 114, 27,
		115, 2, 115, 9,
		115, 20, 115, 27,
		116, 2, 116, 14,
		116, 20, 116, 29,
		117, 2, 117, 16,
		117, 20, 117, 34,
		118, 2, 118, 15,
		118, 20, 118, 33,
		119, 2, 119, 18,
		119, 20, 119, 36,
		120, 2, 120, 9,
		120, 20, 120, 23,
		121, 2, 121, 12,
		121, 20, 121, 34,
		122, 2, 122, 15,
		122, 20, 122, 26,
		123, 2, 123, 10,
		124, 2, 124, 14,
		124, 20, 124, 28,
		129, 6, 129, 7,
		129, 9, 129, 15,
		129, 16, 129, 17,
		129, 20, 129, 25,
		130, 13, 130, 14,
		130, 15, 130, 23,
		138, 10, 138, 11,
		138, 12, 138, 20,
		141, 10, 141, 11,
		141, 12, 141, 26,
		141, 27, 141, 32,
		141, 33, 141, 34,
		141, 35, 141, 42,
		141, 43, 141, 60,
		144, 3, 144, 6,
		144, 10, 144, 11,
		144, 12, 144, 24,
		144, 25, 144, 30,
		145, 3, 145, 11,
		145, 13, 145, 24,
		145, 26, 145, 29,
		145, 33, 145, 34,
		145, 35, 145, 40,
		145, 41, 145, 50,
		145, 51, 145, 52,
		145, 53, 145, 60,
		145, 62, 145, 65,
		146, 6, 146, 9,
		147, 4, 147, 5,
		147, 6, 147, 13,
		147, 14, 147, 21,
		147, 22, 147, 28,
		147, 29, 147, 34,
		148, 5, 148, 8,
		148, 9, 148, 13,
		148, 21, 148, 24,
		149, 5, 149, 8,
		149, 9, 149, 14,
		149, 15, 149, 18,
		153, 3, 153, 14,
		154, 7, 154, 8,
		154, 10, 154, 12,
		154, 22, 154, 30,
		155, 4, 155, 15,
		155, 16, 155, 18,
		158, 7, 158, 8,
		158, 10, 158, 12,
		158, 22, 158, 25,
		159, 7, 159, 8,
		159, 10, 159, 12,
		159, 16, 159, 27,
		159, 28, 159, 30,
		159, 34, 159, 36,
		160, 8, 160, 9,
		160, 10, 160, 22,
		160, 23, 160, 29,
		160, 30, 160, 32,
		161, 6, 161, 7,
		161, 8, 161, 15,
		161, 16, 161, 23,
		161, 24, 161, 30,
		161, 31, 161, 36,
		162, 7, 162, 10,
		162, 11, 162, 14,
		162, 21, 162, 23,
		167, 10, 167, 21,
		168, 4, 168, 5,
		168, 6, 168, 13,
		168, 14, 168, 21,
		168, 22, 168, 28,
		168, 29, 168, 33,
		168, 58, 168, 61,
		168, 62, 168, 66,
		168, 74, 168, 85,
		171, 7, 171, 8,
		171, 10, 171, 12,
		171, 22, 171, 33,
		172, 4, 172, 5,
		172, 6, 172, 18,
		172, 19, 172, 25,
		172, 26, 172, 28,
		177, 5, 177, 17,
		177, 25, 177, 29,
		177, 30, 177, 34,
		178, 4, 178, 5,
		178, 6, 178, 13,
		178, 14, 178, 27,
		179, 2, 179, 14,
		179, 17, 179, 18,
		179, 19, 179, 32,
		179, 33, 179, 38,
		179, 39, 179, 40,
		179, 41, 179, 48,
		179, 49, 179, 62,
		181, 2, 181, 14,
		181, 27, 181, 31,
		181, 32, 181, 36,
		184, 5, 184, 11,
		186, 0, 186, 4,
		188, 5, 188, 6,
		188, 7, 188, 14,
		188, 15, 188, 27,
		188, 36, 188, 37,
		188, 38, 188, 49,
		188, 53, 188, 54,
		188, 55, 188, 62,
		188, 63, 188, 75,
		189, 3, 189, 9,
		190, 9, 190, 13,
		193, 2, 193, 4,
		193, 6, 193, 9,
		193, 13, 193, 14,
		193, 15, 193, 31,
		194, 5, 194, 8,
		198, 6, 198, 7,
		198, 8, 198, 18,
		198, 19, 198, 22,
		198, 35, 198, 41,
		198, 42, 198, 44,
		198, 45, 198, 48,
		198, 50, 198, 51,
		198, 52, 198, 62,
		198, 63, 198, 66,
		200, 10, 200, 14,
		203, 3, 203, 4,
		203, 5, 203, 12,
		203, 13, 203, 20,
		203, 21, 203, 27,
		203, 28, 203, 33,
		204, 4, 204, 7,
		204, 8, 204, 14,
		204, 23, 204, 24,
		204, 25, 204, 32,
		204, 33, 204, 37,
		205, 4, 205, 7,
		205, 8, 205, 13,
		205, 14, 205, 17,
		208, 2, 208, 7,
		208, 11, 208, 12,
		208, 13, 208, 20,
		208, 21, 208, 29,
		209, 5, 209, 7,
		213, 3, 213, 8,
		219, 3, 219, 4,
		219, 5, 219, 16,
		223, 9, 223, 10,
		223, 11, 223, 23,
		223, 24, 223, 29,
		223, 30, 223, 35,
		224, 9, 224, 10,
		224, 11, 224, 21,
		224, 22, 224, 26,
		225, 9, 225, 13,
		226, 9, 226, 21,
		227, 3, 227, 9,
		228, 9, 228, 13,
		232, 1, 232, 2,
		232, 3, 232, 10,
		232, 11, 232, 18,
		232, 19, 232, 25,
		232, 26, 232, 30,
		232, 61, 232, 64,
		232, 65, 232, 71,
		232, 82, 232, 88,
		233, 1, 233, 2,
		233, 3, 233, 5,
		233, 6, 233, 10,
		239, 6, 239, 7,
		239, 9, 239, 15,
		239, 16, 239, 17,
		239, 20, 239, 24,
		240, 1, 240, 2,
		240, 3, 240, 16,
		241, 1, 241, 2,
		241, 3, 241, 7,
		245, 6, 245, 7,
		245, 9, 245, 15,
		245, 16, 245, 17,
		245, 20, 245, 24,
		246, 3, 246, 4,
		246, 5, 246, 13,
		252, 6, 252, 7,
		252, 9, 252, 15,
		252, 16, 252, 17,
		252, 20, 252, 36,
		252, 40, 252, 48,
		252, 55, 252, 58,
		256, 8, 256, 9,
		256, 10, 256, 26,
		257, 8, 257, 9,
		257, 10, 257, 20,
		257, 21, 257, 25,
		258, 16, 258, 17,
		258, 18, 258, 28,
		258, 29, 258, 32,
		261, 6, 261, 14,
		266, 3, 266, 4,
		266, 5, 266, 21,
		270, 1, 270, 12,
		270, 14, 270, 35,
		270, 37, 270, 40,
		270, 44, 270, 45,
		270, 46, 270, 60,
		270, 61, 270, 62,
		270, 63, 270, 73,
		271, 4, 271, 7,
		272, 16, 272, 22,
		272, 23, 272, 27,
		272, 28, 272, 31,
		274, 5, 274, 16,
		280, 1, 280, 7,
		280, 9, 280, 17,
		280, 19, 280, 22,
		280, 26, 280, 27,
		280, 28, 280, 33,
		280, 34, 280, 41,
		280, 42, 280, 43,
		280, 44, 280, 54,
		280, 56, 280, 57,
		280, 58, 280, 65,
		280, 66, 280, 80,
		280, 82, 280, 103,
		281, 4, 281, 7,
		282, 16, 282, 22,
		282, 23, 282, 27,
		282, 28, 282, 31,
		284, 5, 284, 13,
		290, 1, 290, 11,
		290, 13, 290, 30,
		290, 34, 290, 36,
		290, 37, 290, 57,
		292, 2, 292, 8,
		292, 9, 292, 24,
		292, 25, 292, 26,
		292, 27, 292, 34,
		292, 36, 292, 37,
		292, 38, 292, 45,
		292, 46, 292, 53,
		292, 54, 292, 66,
		292, 67, 292, 73,
		293, 2, 293, 3,
		293, 4, 293, 11,
		293, 12, 293, 16,
		295, 1, 295, 10,
		295, 12, 295, 18,
		295, 22, 295, 29,
		295, 30, 295, 40,
		295, 41, 295, 58,
		296, 1, 296, 11,
		296, 15, 296, 20,
		296, 21, 296, 27,
		296, 28, 296, 45,
		296, 47, 296, 48,
		296, 49, 296, 56,
		296, 57, 296, 64,
		296, 65, 296, 71,
		299, 5, 299, 6,
		299, 7, 299, 19,
		299, 20, 299, 23,
		299, 24, 299, 30,
		299, 31, 299, 39,
		299, 43, 299, 49,
		300, 2, 300, 12,
		300, 13, 300, 22,
		300, 23, 300, 28,
		300, 29, 300, 34,
		300, 35, 300, 54,
		301, 2, 301, 12,
		301, 13, 301, 19,
		302, 16, 302, 35,
		306, 1, 306, 2,
		306, 3, 306, 10,
		306, 11, 306, 18,
		306, 19, 306, 26,
		306, 27, 306, 30,
		307, 1, 307, 11,
		307, 12, 307, 16,
		307, 51, 307, 54,
		307, 55, 307, 58,
		307, 65, 307, 71,
		307, 72, 307, 80,
		308, 1, 308, 12,
		308, 16, 308, 27,
		308, 28, 308, 32,
		309, 2, 309, 11,
		309, 15, 309, 20,
		309, 21, 309, 26,
		309, 27, 309, 32,
		309, 33, 309, 36,
		309, 50, 309, 56,
		309, 57, 309, 65,
		312, 4, 312, 8,
		312, 10, 312, 12,
		312, 16, 312, 17,
		312, 18, 312, 25,
		312, 27, 312, 36,
		312, 37, 312, 38,
		312, 42, 312, 44,
		313, 2, 313, 8,
		313, 10, 313, 25,
		313, 27, 313, 41,
		313, 45, 313, 46,
		313, 47, 313, 54,
		313, 55, 313, 62,
		313, 63, 313, 73,
		313, 74, 313, 83,
		313, 84, 313, 88,
		313, 89, 313, 98,
		313, 105, 313, 116,
		315, 2, 315, 6,
		315, 7, 315, 16,
		315, 17, 315, 23,
		315, 25, 315, 40,
		315, 41, 315, 45,
		315, 46, 315, 49,
		315, 50, 315, 59,
		315, 75, 315, 81,
		316, 2, 316, 16,
		316, 20, 316, 31,
		316, 32, 316, 36,
		319, 1, 319, 2,
		319, 3, 319, 5,
		319, 6, 319, 9,
		323, 6, 323, 10,
		323, 12, 323, 14,
		323, 18, 323, 19,
		323, 20, 323, 27,
		323, 29, 323, 38,
		323, 39, 323, 40,
		323, 44, 323, 46,
		328, 4, 328, 11,
		328, 13, 328, 29,
		328, 31, 328, 45,
		328, 49, 328, 50,
		328, 51, 328, 58,
		328, 59, 328, 66,
		328, 67, 328, 77,
		328, 78, 328, 88,
		328, 89, 328, 93,
		328, 94, 328, 111,
		328, 118, 328, 129,
		329, 10, 329, 24,
		329, 28, 329, 39,
		329, 40, 329, 44,
		331, 4, 331, 8,
		331, 9, 331, 19,
		331, 20, 331, 27,
		331, 29, 331, 45,
		331, 46, 331, 50,
		331, 51, 331, 54,
		331, 55, 331, 64,
		331, 81, 331, 87,
		336, 9, 336, 10,
		336, 11, 336, 23,
		336, 24, 336, 30,
		336, 31, 336, 37,
		336, 38, 336, 46,
		337, 3, 337, 4,
		337, 5, 337, 12,
		337, 13, 337, 20,
		337, 21, 337, 28,
		337, 29, 337, 32,
		338, 3, 338, 4,
		338, 5, 338, 21,
		339, 3, 339, 4,
		339, 5, 339, 7,
		339, 8, 339, 12,
		340, 3, 340, 13,
		340, 14, 340, 20,
		343, 5, 343, 8,
		343, 12, 343, 13,
		343, 14, 343, 20,
		343, 21, 343, 30,
		343, 32, 343, 49,
		343, 51, 343, 57,
		343, 60, 343, 63,
		344, 3, 344, 13,
		344, 14, 344, 19,
		344, 49, 344, 52,
		344, 53, 344, 58,
		344, 59, 344, 62,
		353, 6, 353, 7,
		353, 9, 353, 15,
		353, 16, 353, 17,
		353, 20, 353, 26,
		353, 27, 353, 30,
		353, 32, 353, 45,
		353, 46, 353, 53,
		353, 54, 353, 61,
		353, 63, 353, 69,
		353, 70, 353, 71,
		353, 74, 353, 77,
		354, 5, 354, 14,
		355, 1, 355, 4,
		355, 6, 355, 15,
		355, 17, 355, 29,
		355, 33, 355, 34,
		355, 35, 355, 42,
		355, 43, 355, 50,
		355, 51, 355, 61,
		355, 62, 355, 68,
		355, 69, 355, 73,
		355, 74, 355, 77,
		355, 80, 355, 89,
		355, 91, 355, 102,
		355, 103, 355, 107,
		358, 5, 358, 14,
		358, 25, 358, 28,
		359, 3, 359, 12,
		359, 15, 359, 18,
		361, 2, 361, 14,
		361, 18, 361, 29,
		361, 30, 361, 34,
		365, 4, 365, 5,
		365, 6, 365, 13,
		365, 14, 365, 34,
		366, 6, 366, 12,
		366, 13, 366, 20,
		366, 21, 366, 31,
		367, 2, 367, 5,
		367, 7, 367, 13,
		367, 16, 367, 23,
		367, 24, 367, 36,
		367, 37, 367, 40,
		367, 42, 367, 46,
		367, 47, 367, 50,
		367, 53, 367, 56,
		367, 57, 367, 58,
		367, 59, 367, 66,
		367, 67, 367, 87,
		368, 8, 368, 14,
		372, 1, 372, 10,
		372, 13, 372, 14,
		372, 15, 372, 22,
		372, 23, 372, 29,
		372, 30, 372, 33,
		372, 35, 372, 44,
		372, 45, 372, 49,
		372, 50, 372, 53,
		372, 54, 372, 63,
		372, 76, 372, 82,
		374, 4, 374, 5,
		374, 6, 374, 13,
		374, 14, 374, 34,
		374, 42, 374, 48,
		374, 49, 374, 51,
		374, 52, 374, 61,
		374, 63, 374, 70,
		374, 71, 374, 87,
		375, 2, 375, 11,
		375, 14, 375, 20,
		375, 21, 375, 25,
		375, 26, 375, 35,
		375, 37, 375, 40,
		375, 41, 375, 48,
		375, 94, 375, 95,
		375, 96, 375, 103,
		375, 104, 375, 124,
		378, 4, 378, 11,
		378, 12, 378, 26,
		378, 27, 378, 36,
		378, 41, 378, 50,
		378, 61, 378, 62,
		378, 63, 378, 76,
		378, 77, 378, 83,
		378, 84, 378, 92,
		378, 96, 378, 105,
		378, 107, 378, 110,
		378, 111, 378, 114,
		379, 5, 379, 11,
		379, 13, 379, 20,
		379, 24, 379, 25,
		379, 26, 379, 31,
		379, 32, 379, 42,
		379, 43, 379, 56,
		379, 58, 379, 64,
		379, 65, 379, 73,
		379, 77, 379, 86,
		379, 97, 379, 104,
		380, 10, 380, 16,
		380, 17, 380, 21,
		380, 22, 380, 29,
		381, 12, 381, 18,
		382, 3, 382, 12,
		382, 13, 382, 17,
		382, 45, 382, 48,
		382, 49, 382, 54,
		382, 55, 382, 64,
		384, 11, 384, 20,
		385, 5, 385, 11,
		385, 13, 385, 20,
		385, 24, 385, 25,
		385, 26, 385, 31,
		385, 32, 385, 43,
		385, 44, 385, 57,
		385, 59, 385, 65,
		385, 66, 385, 74,
		385, 78, 385, 87,
		385, 98, 385, 105,
		386, 10, 386, 16,
		386, 17, 386, 21,
		386, 22, 386, 29,
		387, 12, 387, 18,
		388, 3, 388, 12,
		388, 13, 388, 17,
		388, 46, 388, 49,
		388, 50, 388, 55,
		388, 56, 388, 65,
		391, 5, 391, 11,
		391, 13, 391, 20,
		391, 24, 391, 25,
		391, 26, 391, 31,
		391, 32, 391, 44,
		391, 45, 391, 58,
		391, 60, 391, 66,
		391, 67, 391, 75,
		391, 80, 391, 87,
		392, 10, 392, 16,
		392, 17, 392, 21,
		392, 22, 392, 29,
		393, 12, 393, 18,
		394, 3, 394, 12,
		394, 13, 394, 18,
		398, 1, 398, 10,
		398, 11, 398, 16,
		405, 6, 405, 7,
		405, 9, 405, 15,
		405, 16, 405, 17,
		405, 20, 405, 33,
		405, 34, 405, 36,
		405, 42, 405, 51,
		405, 53, 405, 59,
		406, 8, 406, 14,
		406, 15, 406, 17,
		406, 18, 406, 27,
		406, 29, 406, 35,
		406, 40, 406, 41,
		406, 42, 406, 54,
		406, 55, 406, 58,
		406, 59, 406, 61,
		406, 67, 406, 73,
		406, 74, 406, 76,
		406, 77, 406, 86,
		406, 88, 406, 95,
		406, 96, 406, 112,
		410, 6, 410, 7,
		410, 9, 410, 15,
		410, 16, 410, 17,
		410, 20, 410, 34,
		410, 35, 410, 38,
		410, 39, 410, 46,
		410, 47, 410, 54,
		410, 57, 410, 68,
		410, 75, 410, 96,
		410, 102, 410, 105,
		411, 4, 411, 5,
		411, 7, 411, 9,
		411, 13, 411, 14,
		411, 15, 411, 22,
		411, 24, 411, 38,
		411, 41, 411, 43,
		412, 9, 412, 10,
		412, 11, 412, 21,
		412, 22, 412, 25,
		412, 27, 412, 28,
		412, 29, 412, 36,
		412, 37, 412, 44,
		412, 45, 412, 51,
	}

	encoded, err := EncodeRanges(ranges)
	if err != nil {
		t.Fatalf("unexpected error encoding ranges: %s", err)
	}

	// Internal decode
	decodedFlattenedRanges, err := DecodeFlattenedRanges(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding ranges: %s", err)
	}
	if diff := cmp.Diff(ranges, decodedFlattenedRanges); diff != "" {
		t.Fatalf("unexpected ranges (-want +got):\n%s", diff)
	}

	// External decode
	decodedSCIPRanges, err := DecodeRanges(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding ranges: %s", err)
	}
	expectedSCIPRanges := make([]*scip.Range, 0, len(ranges)/4)
	for i := 0; i < len(ranges); i += 4 {
		expectedSCIPRanges = append(expectedSCIPRanges, scip.NewRange(ranges[i:i+4]))
	}
	if diff := cmp.Diff(expectedSCIPRanges, decodedSCIPRanges); diff != "" {
		t.Fatalf("unexpected ranges (-want +got):\n%s", diff)
	}
}