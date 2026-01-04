package models

// Structs for parsing the JSON template
type Node struct {
	ClassName string `json:"className"`
	Attrs     Attrs  `json:"attrs"`
	Children  []Node `json:"children,omitempty"`
}

type Attrs struct {
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	ScaleX       float64 `json:"scaleX"`
	ScaleY       float64 `json:"scaleY"`
	Rotation     float64 `json:"rotation"`
	BlurRadius   float64 `json:"blur_radius"`
	Brightness   float64 `json:"brightness"`    // 0 = no change, positive/negative for adjustment
	Contrast     float64 `json:"contrast"`      // 0 = no change, positive/negative for adjustment
	BorderRadius float64 `json:"border_radius"` // in pixels

	// text
	FontSize float64 `json:"fontSize"`
	Text     string  `json:"text"`
	Fill     string  `json:"fill"` // hex color code

	// image
	Path string `json:"path"`
}
