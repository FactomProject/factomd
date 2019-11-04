//+build ignore

//Ͼ/*
// The FactomGenerate templates use Canadian Aboriginal syllabary characters using "Ͼ" U+1438, "ᐳ" U+1433 as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/ᐳ

package templates // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ./generate.go

//Ͼdefine "publish-imports"ᐳ

//Ͼendᐳ

//Ͼdefine "publish"ᐳ
// Start Publish generated go code

func Publish_Ͼ_typeᐳ(parent string, name string, object Ͼ_typeᐳ) Ͼ_typeᐳ {
	return object
}

// End Publish generated go code
//Ͼend ᐳ
