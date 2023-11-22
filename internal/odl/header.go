package odl

type Header struct {
	Magic uint64
	Version uint32
	Capabilities uint32
	ProcessingFlags uint32
	ImagePlatform uint32
	ImageType uint32
	ImageVersion [64]byte
	ImagePlatformVersion [64]byte
	Reserved [100]byte
}

const HeaderMagicValue = 0x44454e4f47464245;

const ImagePlatform_Windows = 0;
const ImagePlatform_Mac = 1;
const ImagePlatform_Windows10X = 2;           // Obsolete, do not use
const ImagePlatform_MacAppleSiliconNative = 3;
const ImagePlatform_MacAppleSiliconRosetta = 4;

const ImageType_Debug = 0;
const ImageType_Ship = 1;
const ImageType_Retail = 2;

const Capabilities_AutoFLF = 0x01;
const Capabilities_64BitPointers = 0x02;
const Capabilities_PrivacyObfuscation = 0x04;
const Capabilities_KernelLogs = 0x08;
const Capabilities_CompressedContents = 0x10;
const Capabilities_CompressedContentsChunked = 0x20;
const Capabilities_TraceID = 0x40;
const Capabilities_PrivacyObfuscationGeneral = 0x80;
