# SMS Code Remote Access

Provides remote access to SMS codes (2FA, banking, etc) sent to your iPhone, via a Mac with Messages in the Cloud enabled.

If you:

* need other people (significant other, virtual assistant, etc) to see 2-factor/banking/etc codes sent over SMS to your iPhone,
* and you have an always-on Mac (like an iMac or Mac Mini),
* and you have Messages in the Cloud enabled,
* and you have a static IP address at home (or some way to connect from outside),

run this daemon, and it'll display a web page with the last 10 SMS messages containing authorization codes sent to your phone.


## Installation

Prerequisites:

* Make sure you have Messages in the Cloud enabled, and see SMS messages in the Messages app on your Mac.

The easy way:

* TODO: make binary releases and/or Homebrew recipe!
* or maybe even build a Mac app...

The hard way:

1. Install Go via Homebrew: `brew install go` (make sure you have Homebrew installed first, of course)

2. Install this app via Go: `go get -u github.com/andreyvit/smscoderemoteaccess`

3. Prepare a JSON config file, see `config-example.json` for example.

4. Make sure Terminal can access your messages database. Open Terminal and run `head -1 ~/Library/Messages/chat.db`; do whatever dance necessary to get that to run (might include adding Terminal.app/iTerm.app to Full Disk Access permissions preference pane),

5. Run `~/go/bin/smscoderemoteaccess -f path/to/your/config.json`.

6. Set up your router to forward whatever port you have configured in config JSON to your Mac.

To use:

1. Open `http://host:port/` in your browser, where `host` is your external IP address (or domain name) and `port` is the port you have configured on your router.

2. Authenticate with the username and password from the config file.
