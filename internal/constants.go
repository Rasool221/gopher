package internal

// fileExtensionTLDs are labels that are technically ICANN-registered TLDs but, in an href, are far
// more likely to be a file extension. We keep refs ending in these as relative paths to the current
// page rather than promoting them to external hosts.
var fileExtensionTLDs = map[string]bool{
	"md":  true, // Markdown vs. Moldova ccTLD
	"sh":  true, // shell script vs. Saint Helena ccTLD
	"zip": true, // archive vs. gTLD
	"mov": true, // QuickTime video vs. gTLD
}
