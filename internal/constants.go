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

var supportedHTMLKeys = map[string]bool{
	"href":   true, // for links
	"src":    true, // for resources
	"poster": true, // for resources (<video /> thumbnail image)
	// TODO: support "srcset"?
}

// socks4 is not supported because the Go standard library doesn't support it.
var supportedProxySchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"socks5": true,
}

var proxyHealthcheckURLs = []string{"http://www.gstatic.com/generate_204", "http://detectportal.firefox.com/success.txt"}

// gopherSqliteDBName is the name of the SQLite DB generated
// with Gopher's output if the output CLI param is set to SQLite.
const gopherSqliteDBName = "gopher_output.db"

// gopherSqliteDBCreateSQL is the SQL that's ran on the SQLite
// DB to initialize it if the user selectedop
const gopherSqliteDBCreateSQL = `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id UUID NOT NULL,
		url TEXT NOT NULL,
		parent_url TEXT
	);

	CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id UUID NOT NULL,
		resource_url TEXT NOT NULL,
		file_type TEXT NOT NULL,
		parent_url TEXT
	);

	CREATE TABLE IF NOT EXISTS errors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id UUID NOT NULL,
		error_message TEXT NOT NULL,
		parent_url TEXT
	);
`

// gopherSqliteInsertLinkSQL is the SQL that inserts a link into the SQLite DB.
const gopherSqliteInsertLinkSQL = `
	INSERT INTO urls (run_id, url, parent_url) VALUES (?, ?, ?);
`

// gopherSqliteInsertResourceSQL is the SQL that inserts a resource into the SQLite DB.
const gopherSqliteInsertResourceSQL = `
  INSERT INTO resources (run_id, resource_url, file_type, parent_url) VALUES (?, ?, ?, ?)
`

// gopherSqliteInsertErrorSQL is the SQL that inserts an error into the SQLite DB.
const gopherSqliteInsertErrorSQL = `
  INSERT INTO errors (run_id, error_message, parent_url) VALUES (?, ?, ?)
`
