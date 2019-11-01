//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and ' ᐳ' so ᐸ_typenameᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

package generated // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ./generate.go

//ᐸdefine "publish-imports"ᐳ

//ᐸendᐳ

//ᐸdefine "publish"ᐳ
// Start Publish generated go code

func Publish_ᐸ_typeᐳ(parent string, name string, object ᐸ_typeᐳ) ᐸ_typeᐳ {
	return object
}

// End Publish generated go code
//ᐸend ᐳ
