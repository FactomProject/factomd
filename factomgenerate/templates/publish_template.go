//+build ignore

//Ͼ/*
// The FactomGenerate templates use Greek Capitol  syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package templates // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ./generate.go

//Ͼdefine "publish-imports"Ͽ

//ϾendϿ

//Ͼdefine "publish"Ͽ
// Start Publish generated go code

func Publish_Ͼ_typeϿ(parent string, name string, object Ͼ_typeϿ) Ͼ_typeϿ {
	return object
}

// End Publish generated go code
//Ͼend Ͽ
