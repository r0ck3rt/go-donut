package donut

import (
	"bytes"

	"github.com/google/uuid"
)

const (
	DONUT_MAX_PARAM   = 8 // maximum number of parameters passed to method
	DONUT_MAX_NAME    = 256
	DONUT_MAX_DLL     = 8 // maximum number of DLL supported by instance
	DONUT_MAX_URL     = 256
	DONUT_MAX_MODNAME = 8
	DONUT_SIG_LEN     = 8 // 64-bit string to verify decryption ok
	DONUT_VER_LEN     = 32
	DONUT_DOMAIN_LEN  = 8

	MARU_MAX_STR  = 64
	MARU_BLK_LEN  = 16
	MARU_HASH_LEN = 8
	MARU_IV_LEN   = 8

	DONUT_RUNTIME_NET4 = "v4.0.30319"

	NTDLL_DLL    = "ntdll.dll"
	KERNEL32_DLL = "kernel32.dll"
	ADVAPI32_DLL = "advapi32.dll"
	CRYPT32_DLL  = "crypt32.dll"
	MSCOREE_DLL  = "mscoree.dll"
	OLE32_DLL    = "ole32.dll"
	OLEAUT32_DLL = "oleaut32.dll"
	WININET_DLL  = "wininet.dll"
	COMBASE_DLL  = "combase.dll"
	USER32_DLL   = "user32.dll"
	SHLWAPI_DLL  = "shlwapi.dll"
)

// DonutArch - CPU architecture type (32, 64, or 32+64)
type DonutArch int

const (
	// X32 - 32bit
	X32 DonutArch = iota
	// X64 - 64 bit
	X64
	// X84 - 32+64 bit
	X84
)

type ModuleType int

const (
	DONUT_MODULE_NET_DLL ModuleType = iota // .NET DLL. Requires class and method
	DONUT_MODULE_NET_EXE                   // .NET EXE. Executes Main if no class and method provided
	DONUT_MODULE_DLL                       // Unmanaged DLL, function is optional
	DONUT_MODULE_EXE                       // Unmanaged EXE
	DONUT_MODULE_VBS                       // VBScript
	DONUT_MODULE_JS                        // JavaScript or JScript
	DONUT_MODULE_XSL                       // XSL with JavaScript/JScript or VBscript embedded
)

type InstanceType int

const (
	DONUT_INSTANCE_PIC InstanceType = 1 // Self-contained
	DONUT_INSTANCE_URL              = 2 // Download from remote server
)

type DonutConfig struct {
	Arch       DonutArch
	Type       ModuleType
	InstType   InstanceType
	Parameters string // separated by , or ;

	NoCrypto bool

	Domain  string // .NET stuff
	Class   string
	Method  string // Used by Native DLL and .NET DLL
	Runtime string

	Module     *DonutModule
	ModuleName string
	URL        string
	ModuleMac  uint64
	ModuleData *bytes.Buffer

	inst    *DonutInstance
	instLen uint32
}

type DonutModule struct {
	Type       uint32                                  // EXE, DLL, JS, VBS, XSL
	Runtime    [DONUT_MAX_NAME]uint16                  // runtime version for .NET EXE/DLL
	Domain     [DONUT_MAX_NAME]uint16                  // domain name to use for .NET EXE/DLL
	Cls        [DONUT_MAX_NAME]uint16                  // name of class and optional namespace for .NET EXE/DLL
	Method     [DONUT_MAX_NAME * 2]byte                // name of method to invoke for .NET DLL or api for unmanaged DLL
	ParamCount uint32                                  // number of parameters for DLL/EXE
	Param      [DONUT_MAX_PARAM][DONUT_MAX_NAME]uint16 // string parameters for DLL/EXE
	Sig        [DONUT_MAX_NAME]byte                    // random string to verify decryption
	Mac        uint64                                  // to verify decryption was ok
	Len        uint64                                  // size of EXE/DLL/XSL/JS/VBS file
	Data       [4]byte                                 // data of EXE/DLL/XSL/JS/VBS file
}

type DonutCrypt struct {
	Mk  [CipherKeyLen]byte   // master key
	Ctr [CipherBlockLen]byte // counter + nonce
}

type DonutInstance struct {
	Len  uint32     // total size of instance
	Key  DonutCrypt // decrypts instance
	Iv   [8]byte    // the 64-bit initial value for maru hash
	Hash [64]uint64 // holds up to 64 api hashes/addrs {api}

	// everything from here is encrypted
	ApiCount int                     // the 64-bit hashes of API required for instance to work
	DllCount int                     // the number of DLL to load before resolving API
	DllName  [DONUT_MAX_DLL][32]byte // a list of DLL strings to load

	s [8]byte // amsi.dll

	bypass         int      // indicates behaviour of byassing AMSI/WLDP
	clr            [8]byte  // clr.dll
	wldp           [16]byte // wldp.dll
	wldpQuery      [32]byte // WldpQueryDynamicCodeTrust
	wldpIsApproved [32]byte // WldpIsClassInApprovedList
	amsiInit       [16]byte // AmsiInitialize
	amsiScanBuf    [16]byte // AmsiScanBuffer
	amsiScanStr    [16]byte // AmsiScanString

	wscript     [8]uint16  // WScript
	wscript_exe [16]uint16 // wscript.exe

	xIID_IUnknown  uuid.UUID
	xIID_IDispatch uuid.UUID

	//  GUID required to load .NET assemblies
	xCLSID_CLRMetaHost    uuid.UUID
	xIID_ICLRMetaHost     uuid.UUID
	xIID_ICLRRuntimeInfo  uuid.UUID
	xCLSID_CorRuntimeHost uuid.UUID
	xIID_ICorRuntimeHost  uuid.UUID
	xIID_AppDomain        uuid.UUID

	//  GUID required to run VBS and JS files
	xCLSID_ScriptLanguage        uuid.UUID // vbs or js
	xIID_IHost                   uuid.UUID // wscript object
	xIID_IActiveScript           uuid.UUID // engine
	xIID_IActiveScriptSite       uuid.UUID // implementation
	xIID_IActiveScriptSiteWindow uuid.UUID // basic GUI stuff
	xIID_IActiveScriptParse32    uuid.UUID // parser
	xIID_IActiveScriptParse64    uuid.UUID

	//  GUID required to run XSL files
	xCLSID_DOMDocument30 uuid.UUID
	xIID_IXMLDOMDocument uuid.UUID
	xIID_IXMLDOMNode     uuid.UUID

	Type int // DONUT_INSTANCE_PIC or DONUT_INSTANCE_URL

	Url [DONUT_MAX_URL]byte // staging server hosting donut module
	Req [8]byte             // just a buffer for "GET"

	sig [DONUT_MAX_NAME]byte // string to hash
	mac uint64               // to verify decryption ok

	mod_key DonutCrypt // used to decrypt module
	mod_len uint64     // total size of module

	/*  union {
	    PDONUT_MODULE p;         // for URL
	    DONUT_MODULE  x;         // for PIC
	  } module; */
}

type API_IMPORT struct {
	Module string
	Name   string
}

var api_imports = []API_IMPORT{
	API_IMPORT{Module: KERNEL32_DLL, Name: "LoadLibraryA"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "GetProcAddress"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "GetModuleHandleA"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "VirtualAlloc"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "VirtualFree"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "VirtualQuery"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "VirtualProtect"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "Sleep"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "MultiByteToWideChar"},
	API_IMPORT{Module: KERNEL32_DLL, Name: "GetUserDefaultLCID"},

	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayCreate"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayCreateVector"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayPutElement"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayDestroy"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayGetLBound"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SafeArrayGetUBound"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SysAllocString"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "SysFreeString"},
	API_IMPORT{Module: OLEAUT32_DLL, Name: "LoadTypeLib"},

	API_IMPORT{Module: WININET_DLL, Name: "InternetCrackUrlA"},
	API_IMPORT{Module: WININET_DLL, Name: "InternetOpenA"},
	API_IMPORT{Module: WININET_DLL, Name: "InternetConnectA"},
	API_IMPORT{Module: WININET_DLL, Name: "InternetSetOptionA"},
	API_IMPORT{Module: WININET_DLL, Name: "InternetReadFile"},
	API_IMPORT{Module: WININET_DLL, Name: "InternetCloseHandle"},
	API_IMPORT{Module: WININET_DLL, Name: "HttpOpenRequestA"},
	API_IMPORT{Module: WININET_DLL, Name: "HttpSendRequestA"},
	API_IMPORT{Module: WININET_DLL, Name: "HttpQueryInfoA"},

	API_IMPORT{Module: MSCOREE_DLL, Name: "CorBindToRuntime"},
	API_IMPORT{Module: MSCOREE_DLL, Name: "CLRCreateInstance"},

	API_IMPORT{Module: OLE32_DLL, Name: "CoInitializeEx"},
	API_IMPORT{Module: OLE32_DLL, Name: "CoCreateInstance"},
	API_IMPORT{Module: OLE32_DLL, Name: "CoUninitialize"},
}

// required to load .NET assemblies
var ( //todo: the first 6 bytes of these were int32+int16, might need to be swapped
	xCLSID_CorRuntimeHost = uuid.UUID{
		0xcb, 0x2f, 0x67, 0x23, 0xab, 0x3a, 0x11, 0xd2, 0x9c, 0x40, 0x00, 0xc0, 0x4f, 0xa3, 0x0a, 0x3e}

	xIID_ICorRuntimeHost = uuid.UUID{
		0xcb, 0x2f, 0x67, 0x22, 0xab, 0x3a, 0x11, 0xd2, 0x9c, 0x40, 0x00, 0xc0, 0x4f, 0xa3, 0x0a, 0x3e}

	xCLSID_CLRMetaHost = uuid.UUID{
		0x92, 0x80, 0x18, 0x8d, 0x0e, 0x8e, 0x48, 0x67, 0xb3, 0xc, 0x7f, 0xa8, 0x38, 0x84, 0xe8, 0xde}

	xIID_ICLRMetaHost = uuid.UUID{
		0xD3, 0x32, 0xDB, 0x9E, 0xB9, 0xB3, 0x41, 0x25, 0x82, 0x07, 0xA1, 0x48, 0x84, 0xF5, 0x32, 0x16}

	xIID_ICLRRuntimeInfo = uuid.UUID{
		0xBD, 0x39, 0xD1, 0xD2, 0xBA, 0x2F, 0x48, 0x6a, 0x89, 0xB0, 0xB4, 0xB0, 0xCB, 0x46, 0x68, 0x91}

	xIID_AppDomain = uuid.UUID{
		0x05, 0xF6, 0x96, 0xDC, 0x2B, 0x29, 0x36, 0x63, 0xAD, 0x8B, 0xC4, 0x38, 0x9C, 0xF2, 0xA7, 0x13}

	// required to load VBS and JS files
	xIID_IUnknown = uuid.UUID{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}

	xIID_IDispatch = uuid.UUID{
		0x00, 0x02, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}

	xIID_IHost = uuid.UUID{
		0x91, 0xaf, 0xbd, 0x1b, 0x5f, 0xeb, 0x43, 0xf5, 0xb0, 0x28, 0xe2, 0xca, 0x96, 0x06, 0x17, 0xec}

	xIID_IActiveScript = uuid.UUID{
		0xbb, 0x1a, 0x2a, 0xe1, 0xa4, 0xf9, 0x11, 0xcf, 0x8f, 0x20, 0x00, 0x80, 0x5f, 0x2c, 0xd0, 0x64}

	xIID_IActiveScriptSite = uuid.UUID{
		0xdb, 0x01, 0xa1, 0xe3, 0xa4, 0x2b, 0x11, 0xcf, 0x8f, 0x20, 0x00, 0x80, 0x5f, 0x2c, 0xd0, 0x64}

	xIID_IActiveScriptSiteWindow = uuid.UUID{
		0xd1, 0x0f, 0x67, 0x61, 0x83, 0xe9, 0x11, 0xcf, 0x8f, 0x20, 0x00, 0x80, 0x5f, 0x2c, 0xd0, 0x64}

	xIID_IActiveScriptParse32 = uuid.UUID{
		0xbb, 0x1a, 0x2a, 0xe2, 0xa4, 0xf9, 0x11, 0xcf, 0x8f, 0x20, 0x00, 0x80, 0x5f, 0x2c, 0xd0, 0x64}

	xIID_IActiveScriptParse64 = uuid.UUID{
		0xc7, 0xef, 0x76, 0x58, 0xe1, 0xee, 0x48, 0x0e, 0x97, 0xea, 0xd5, 0x2c, 0xb4, 0xd7, 0x6d, 0x17}

	xCLSID_VBScript = uuid.UUID{
		0xB5, 0x4F, 0x37, 0x41, 0x5B, 0x07, 0x11, 0xcf, 0xA4, 0xB0, 0x00, 0xAA, 0x00, 0x4A, 0x55, 0xE8}

	xCLSID_JScript = uuid.UUID{
		0xF4, 0x14, 0xC2, 0x60, 0x6A, 0xC0, 0x11, 0xCF, 0xB6, 0xD1, 0x00, 0xAA, 0x00, 0xBB, 0xBB, 0x58}

	// required to load XSL files
	xCLSID_DOMDocument30 = uuid.UUID{
		0xf5, 0x07, 0x8f, 0x32, 0xc5, 0x51, 0x11, 0xd3, 0x89, 0xb9, 0x00, 0x00, 0xf8, 0x1f, 0xe2, 0x21}

	xIID_IXMLDOMDocument = uuid.UUID{
		0x29, 0x33, 0xBF, 0x81, 0x7B, 0x36, 0x11, 0xD2, 0xB2, 0x0E, 0x00, 0xC0, 0x4F, 0x98, 0x3E, 0x60}

	xIID_IXMLDOMNode = uuid.UUID{
		0x29, 0x33, 0xbf, 0x80, 0x7b, 0x36, 0x11, 0xd2, 0xb2, 0x0e, 0x00, 0xc0, 0x4f, 0x98, 0x3e, 0x60}
)
