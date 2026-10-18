package export

// Minimal PresentationML packaging, authored against the public OOXML format.
// This file uses only Go's standard library, not an Office SDK or a template.
import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

const pptxOfficeRels = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/"
const pptxNamespaces = `xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`

type pptxArchive struct {
	zip           *zip.Writer
	width, height int32
}

type pptxRelationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}

type pptxRelationships struct {
	XMLName xml.Name           `xml:"Relationships"`
	XMLNS   string             `xml:"xmlns,attr"`
	Items   []pptxRelationship `xml:"Relationship"`
}

func (p *pptxArchive) relations(name string, items ...pptxRelationship) error {
	data, err := xml.Marshal(pptxRelationships{XMLNS: "http://schemas.openxmlformats.org/package/2006/relationships", Items: items})
	if err != nil {
		return err
	}
	return p.part(name, append([]byte(xml.Header), data...), zip.Deflate)
}

func pptxRelation(id, kind, target string) pptxRelationship {
	return pptxRelationship{ID: id, Type: pptxOfficeRels + kind, Target: target}
}

func newPPTXArchive(out io.Writer, width, height int32, title string, count int) (*pptxArchive, error) {
	p := &pptxArchive{zip: zip.NewWriter(out), width: width, height: height}
	if err := p.start(title, count); err != nil {
		_ = p.zip.Close()
		return nil, err
	}
	return p, nil
}

func (p *pptxArchive) start(title string, count int) error {
	if err := p.relations("_rels/.rels",
		pptxRelation("rIdOffice", "officeDocument", "ppt/presentation.xml"),
		pptxRelationship{ID: "rIdCore", Type: "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", Target: "docProps/core.xml"},
		pptxRelation("rIdApp", "extended-properties", "docProps/app.xml")); err != nil {
		return err
	}
	if err := p.contentTypes(count); err != nil {
		return err
	}
	var slideIDs bytes.Buffer
	rels := []pptxRelationship{
		pptxRelation("rIdMaster", "slideMaster", "slideMasters/slideMaster1.xml"),
		pptxRelation("rIdNotesMaster", "notesMaster", "notesMasters/notesMaster1.xml"),
		pptxRelation("rIdPresProps", "presProps", "presProps.xml"),
	}
	for i := 1; i <= count; i++ {
		fmt.Fprintf(&slideIDs, `<p:sldId id="%d" r:id="rIdSlide%d"/>`, 255+i, i)
		rels = append(rels, pptxRelation(fmt.Sprintf("rIdSlide%d", i), "slide", fmt.Sprintf("slides/slide%d.xml", i)))
	}
	if err := p.relations("ppt/_rels/presentation.xml.rels", rels...); err != nil {
		return err
	}
	parts := []struct{ name, body string }{
		{"docProps/core.xml", `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>` + pptxEscape(title) + `</dc:title><dc:creator>Animation</dc:creator><dc:description>Animation presentation.</dc:description></cp:coreProperties>`},
		{"docProps/app.xml", fmt.Sprintf(`<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>Animation</Application><PresentationFormat>Custom</PresentationFormat><Slides>%d</Slides><Notes>%d</Notes><HiddenSlides>0</HiddenSlides><ScaleCrop>false</ScaleCrop><LinksUpToDate>false</LinksUpToDate><SharedDoc>false</SharedDoc><HyperlinksChanged>false</HyperlinksChanged></Properties>`, count, count)},
		{"ppt/presentation.xml", fmt.Sprintf(`<p:presentation %s><p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rIdMaster"/></p:sldMasterIdLst><p:notesMasterIdLst><p:notesMasterId r:id="rIdNotesMaster"/></p:notesMasterIdLst><p:sldIdLst>%s</p:sldIdLst><p:sldSz cx="%d" cy="%d" type="custom"/><p:notesSz cx="6858000" cy="9144000"/><p:defaultTextStyle/></p:presentation>`, pptxNamespaces, slideIDs.String(), p.width, p.height)},
		{"ppt/presProps.xml", `<p:presentationPr ` + pptxNamespaces + `/>`},
		{"ppt/slideMasters/slideMaster1.xml", `<p:sldMaster ` + pptxNamespaces + `><p:cSld name="Animation"><p:bg><p:bgPr><a:solidFill><a:srgbClr val="FFFFFF"/></a:solidFill><a:effectLst/></p:bgPr></p:bg><p:spTree>` + pptxGroup + `</p:spTree></p:cSld>` + pptxColorMap + `<p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rIdLayout"/></p:sldLayoutIdLst><p:txStyles><p:titleStyle/><p:bodyStyle/><p:otherStyle/></p:txStyles></p:sldMaster>`},
		{"ppt/slideLayouts/slideLayout1.xml", `<p:sldLayout ` + pptxNamespaces + ` type="blank" preserve="1"><p:cSld name="Blank"><p:spTree>` + pptxGroup + `</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sldLayout>`},
		{"ppt/notesMasters/notesMaster1.xml", `<p:notesMaster ` + pptxNamespaces + `><p:cSld name="Notes"><p:spTree>` + pptxGroup + `</p:spTree></p:cSld>` + pptxColorMap + `<p:notesStyle><a:lvl1pPr><a:defRPr sz="1200"/></a:lvl1pPr></p:notesStyle></p:notesMaster>`},
		{"ppt/theme/theme1.xml", pptxTheme},
		// PowerPoint rejects a package when slide and notes masters share the
		// same theme part, even though generic OOXML schema validation passes.
		{"ppt/theme/theme2.xml", pptxTheme},
	}
	for _, part := range parts {
		if err := p.xmlPart(part.name, part.body); err != nil {
			return err
		}
	}
	if err := p.relations("ppt/slideMasters/_rels/slideMaster1.xml.rels",
		pptxRelation("rIdLayout", "slideLayout", "../slideLayouts/slideLayout1.xml"),
		pptxRelation("rIdTheme", "theme", "../theme/theme1.xml")); err != nil {
		return err
	}
	if err := p.relations("ppt/slideLayouts/_rels/slideLayout1.xml.rels", pptxRelation("rIdMaster", "slideMaster", "../slideMasters/slideMaster1.xml")); err != nil {
		return err
	}
	return p.relations("ppt/notesMasters/_rels/notesMaster1.xml.rels", pptxRelation("rIdTheme", "theme", "../theme/theme2.xml"))
}

func (p *pptxArchive) addFrame(index int, at float64, data []byte) error {
	n := index + 1
	if err := p.part(fmt.Sprintf("ppt/media/frame%d.png", n), data, zip.Store); err != nil {
		return err
	}
	description := pptxEscape(fmt.Sprintf("Animation time: %g seconds", at))
	slide := fmt.Sprintf(`<p:sld %s><p:cSld name="Frame %d"><p:spTree>%s<p:pic><p:nvPicPr><p:cNvPr id="2" name="Animation frame %d" descr="%s"/><p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr><p:blipFill><a:blip r:embed="rIdImage"/><a:stretch><a:fillRect/></a:stretch></p:blipFill><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:ln><a:noFill/></a:ln></p:spPr></p:pic></p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>`, pptxNamespaces, n, pptxGroup, n, description, p.width, p.height)
	if err := p.xmlPart(fmt.Sprintf("ppt/slides/slide%d.xml", n), slide); err != nil {
		return err
	}
	if err := p.relations(fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", n),
		pptxRelation("rIdLayout", "slideLayout", "../slideLayouts/slideLayout1.xml"),
		pptxRelation("rIdImage", "image", fmt.Sprintf("../media/frame%d.png", n)),
		pptxRelation("rIdNotes", "notesSlide", fmt.Sprintf("../notesSlides/notesSlide%d.xml", n))); err != nil {
		return err
	}
	return p.addNotes(n, description)
}

func (p *pptxArchive) addNotes(n int, escapedDescription string) error {
	notes := `<p:notes ` + pptxNamespaces + `><p:cSld><p:spTree>` + pptxGroup + `<p:sp><p:nvSpPr><p:cNvPr id="2" name="Notes"/><p:cNvSpPr txBox="1"/><p:nvPr><p:ph type="body" idx="1"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="914400" y="914400"/><a:ext cx="5029200" cy="7315200"/></a:xfrm></p:spPr><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr lang="en-US" sz="1200"/><a:t>` + escapedDescription + `</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:notes>`
	if err := p.xmlPart(fmt.Sprintf("ppt/notesSlides/notesSlide%d.xml", n), notes); err != nil {
		return err
	}
	return p.relations(fmt.Sprintf("ppt/notesSlides/_rels/notesSlide%d.xml.rels", n),
		pptxRelation("rIdSlide", "slide", fmt.Sprintf("../slides/slide%d.xml", n)),
		pptxRelation("rIdNotesMaster", "notesMaster", "../notesMasters/notesMaster1.xml"))
}

func (p *pptxArchive) xmlPart(name, body string) error {
	return p.part(name, []byte(xml.Header+body), zip.Deflate)
}

func (p *pptxArchive) part(name string, data []byte, method uint16) error {
	header := &zip.FileHeader{Name: name, Method: method, Modified: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)}
	w, err := p.zip.CreateHeader(header)
	if err == nil {
		var n int
		n, err = w.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
	}
	if err != nil {
		return fmt.Errorf("write PPTX part %s: %w", name, err)
	}
	return nil
}

func pptxEscape(text string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(text))
	return out.String()
}

func (p *pptxArchive) contentTypes(count int) error {
	const presentationType = "application/vnd.openxmlformats-officedocument.presentationml."
	var out bytes.Buffer
	out.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/>`)
	for _, part := range []struct{ name, kind string }{
		{"/docProps/core.xml", "application/vnd.openxmlformats-package.core-properties+xml"},
		{"/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml"},
		{"/ppt/presentation.xml", presentationType + "presentation.main+xml"},
		{"/ppt/presProps.xml", presentationType + "presProps+xml"},
		{"/ppt/slideMasters/slideMaster1.xml", presentationType + "slideMaster+xml"},
		{"/ppt/slideLayouts/slideLayout1.xml", presentationType + "slideLayout+xml"},
		{"/ppt/notesMasters/notesMaster1.xml", presentationType + "notesMaster+xml"},
		{"/ppt/theme/theme1.xml", "application/vnd.openxmlformats-officedocument.theme+xml"},
		{"/ppt/theme/theme2.xml", "application/vnd.openxmlformats-officedocument.theme+xml"},
	} {
		fmt.Fprintf(&out, `<Override PartName="%s" ContentType="%s"/>`, part.name, part.kind)
	}
	for n := 1; n <= count; n++ {
		fmt.Fprintf(&out, `<Override PartName="/ppt/slides/slide%d.xml" ContentType="%sslide+xml"/><Override PartName="/ppt/notesSlides/notesSlide%d.xml" ContentType="%snotesSlide+xml"/>`, n, presentationType, n, presentationType)
	}
	out.WriteString(`</Types>`)
	return p.xmlPart("[Content_Types].xml", out.String())
}

const pptxGroup = `<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>`
const pptxColorMap = `<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>`

// Theme defaults are not used to redraw the frame: the displayed content is a
// lossless PNG. A complete theme keeps the package usable by Office clients.
const pptxTheme = `<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Animation"><a:themeElements>
<a:clrScheme name="Animation"><a:dk1><a:srgbClr val="000000"/></a:dk1><a:lt1><a:srgbClr val="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="323232"/></a:dk2><a:lt2><a:srgbClr val="F5F5F5"/></a:lt2><a:accent1><a:srgbClr val="C86E8B"/></a:accent1><a:accent2><a:srgbClr val="91A28E"/></a:accent2><a:accent3><a:srgbClr val="BDA477"/></a:accent3><a:accent4><a:srgbClr val="B28FBA"/></a:accent4><a:accent5><a:srgbClr val="72A0A5"/></a:accent5><a:accent6><a:srgbClr val="BFA698"/></a:accent6><a:hlink><a:srgbClr val="0000FF"/></a:hlink><a:folHlink><a:srgbClr val="800080"/></a:folHlink></a:clrScheme>
<a:fontScheme name="Animation"><a:majorFont><a:latin typeface="Arial"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont><a:minorFont><a:latin typeface="Arial"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont></a:fontScheme>
<a:fmtScheme name="Animation"><a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst><a:lnStyleLst><a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln><a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln><a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln></a:lnStyleLst><a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst><a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst></a:fmtScheme></a:themeElements></a:theme>`
