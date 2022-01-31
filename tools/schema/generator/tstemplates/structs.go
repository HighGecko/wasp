package tstemplates

var structsTs = map[string]string{
	// *******************************
	"structs.ts": `
$#emit importWasmLib
$#emit importWasmTypes
$#each structs structType
`,
	// *******************************
	"structType": `

export class $StrName {
$#each struct structField

    static fromBytes(buf: u8[]|null): $StrName {
        const dec = new wasmtypes.WasmDecoder(buf==null ? [] : buf);
        const data = new $StrName();
$#each struct structDecode
        dec.close();
        return data;
    }

    bytes(): u8[] {
        const enc = new wasmtypes.WasmEncoder();
$#each struct structEncode
        return enc.buf();
    }
}
$#set mut Immutable
$#emit structMethods
$#set mut Mutable
$#emit structMethods
`,
	// *******************************
	"structField": `
    $fldName$fldPad : $fldLangType = $fldTypeInit; $fldComment
`,
	// *******************************
	"structDecode": `
        data.$fldName$fldPad = wasmtypes.$fldType$+Decode(dec);
`,
	// *******************************
	"structEncode": `
		    wasmtypes.$fldType$+Encode(enc, this.$fldName);
`,
	// *******************************
	"structMethods": `

export class $mut$StrName extends wasmtypes.ScProxy {
$#if mut structMethodDelete

    exists(): bool {
        return this.proxy.exists();
    }
$#if mut structMethodSetValue

    value(): $StrName {
        return $StrName.fromBytes(this.proxy.get());
    }
}
`,
	// *******************************
	"structMethodDelete": `

    delete(): void {
        this.proxy.delete();
    }
`,
	// *******************************
	"structMethodSetValue": `

    setValue(value: $StrName): void {
        this.proxy.set(value.bytes());
    }
`,
}
