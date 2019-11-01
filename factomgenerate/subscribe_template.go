//+build ignore
//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and ' ᐳ' so ᐸ_typename ᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

package main // this is only here to make gofmt happy and is never in the generated code

//ᐸdefine "subscribe-imports" ᐳ

//ᐸendᐳ

//ᐸdefine "subscribe" ᐳ
// Start Subscribe generated go code

func Subscribe_ᐸ_typeᐳ(parent string, name string) *ᐸ_type ᐳ {
	return new(ᐸ_typeᐳ)
}
// End Subscribe generated code
//ᐸend ᐳ
