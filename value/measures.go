package value

import "math"

const (
	Metre      Measure = 1
	Decimetre  Measure = 0.1
	Centimetre Measure = 0.01
	Millimetre Measure = 0.001
	Kilometre  Measure = 1000

	Yard Measure = 0.9144
	Inch Measure = 0.0254
	Mile Measure = 1609.34
	Foot Measure = 0.3048
)

const (
	Metre2      Measure = 1
	Decimetre2  Measure = 0.01
	Centimetre2 Measure = 0.0001
	Millimetre2 Measure = 1e-6
	Kilometre2  Measure = 1e+6

	Yard2 Measure = 0.83612736
	Inch2 Measure = 0.00064516
	Mile2 Measure = 2.59e+6
	Foot2 Measure = 0.09290304
)

const (
	Radian Measure = 1
	Degree Measure = (2 * math.Pi) / 360
)

type Measure float64

func (m Measure) Float64() float64 {
	return float64(m)
}

type FigureMeasures struct {
	Length    Measure `json:"length"`
	Perimeter Measure `json:"perimeter"`
	Area      Measure `json:"area"`
	Angle     Measure `json:"angle"`
}

func (fm *FigureMeasures) ToFigureMeasuresNames() *FigureMeasuresNames {
	fmn := NewFigureMeasuresNames()
	fmn.Length = NameOfLengthMeasure(fm.Length)
	fmn.Perimeter = NameOfLengthMeasure(fm.Perimeter)
	fmn.Area = NameOfAreaMeasure(fm.Area)
	fmn.Angle = NameOfAngleMeasure(fm.Angle)
	return fmn
}

func NewFigureMeasures() *FigureMeasures {
	return &FigureMeasures{
		Angle:     Degree,
		Area:      Metre2,
		Length:    Centimetre,
		Perimeter: Metre,
	}
}

type FigureMeasuresNames struct {
	Length    string `json:"length"`
	Area      string `json:"area"`
	Perimeter string `json:"perimeter"`
	Angle     string `json:"angle"`
}

func NewFigureMeasuresNames() *FigureMeasuresNames {
	return &FigureMeasuresNames{
		Length:    "cm",
		Area:      "m2",
		Perimeter: "m",
		Angle:     "deg",
	}
}

func (mn *FigureMeasuresNames) ToFigureMeasures(previousFM *FigureMeasures) *FigureMeasures {
	fm := NewFigureMeasures()
	if previousFM != nil {
		fm.Length = previousFM.Length
		fm.Perimeter = previousFM.Perimeter
		fm.Area = previousFM.Area
		fm.Angle = previousFM.Angle
	}
	if m := LengthMeasureByName(mn.Length); m != 0 {
		fm.Length = m
	}
	if m := AreaMeasureByName(mn.Area); m != 0 {
		fm.Area = m
	}
	if m := LengthMeasureByName(mn.Perimeter); m != 0 {
		fm.Perimeter = m
	}
	if m := AngleMeasureByName(mn.Angle); m != 0 {
		fm.Angle = m
	}
	return fm
}

func NameOfLengthMeasure(m Measure) string {
	switch m {
	case Metre:
		return "m"
	case Decimetre:
		return "dm"
	case Centimetre:
		return "cm"
	case Millimetre:
		return "mm"
	case Kilometre:
		return "km"
	case Yard:
		return "yd"
	case Inch:
		return "in"
	case Mile:
		return "mi"
	case Foot:
		return "ft"
	}
	return ""
}

func NameOfAreaMeasure(m Measure) string {
	switch m {
	case Metre2:
		return "m2"
	case Decimetre2:
		return "dm2"
	case Centimetre2:
		return "cm2"
	case Millimetre2:
		return "mm2"
	case Kilometre2:
		return "km2"
	case Yard2:
		return "yd2"
	case Inch2:
		return "in2"
	case Mile2:
		return "mi2"
	case Foot2:
		return "ft2"
	}
	return ""
}

func NameOfAngleMeasure(m Measure) string {
	switch m {
	case Degree:
		return "deg"
	case Radian:
		return "rad"
	}
	return ""
}

func LengthMeasureByName(name string) Measure {
	switch name {
	case "m":
		return Metre
	case "dm":
		return Decimetre
	case "cm":
		return Centimetre
	case "mm":
		return Millimetre
	case "km":
		return Kilometre
	case "yd":
		return Yard
	case "in":
		return Inch
	case "mi":
		return Mile
	case "ft":
		return Foot
	}
	return 0
}

func AreaMeasureByName(name string) Measure {
	switch name {
	case "m2":
		return Metre2
	case "dm2":
		return Decimetre2
	case "cm2":
		return Centimetre2
	case "mm2":
		return Millimetre2
	case "km2":
		return Kilometre2
	case "yd2":
		return Yard2
	case "in2":
		return Inch2
	case "mi2":
		return Mile2
	case "ft2":
		return Foot2
	}
	return 0
}

func AngleMeasureByName(name string) Measure {
	switch name {
	case "deg":
		return Degree
	case "rad":
		return Radian
	}
	return 0
}
