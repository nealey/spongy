Spongy Server Installation
==========================

You gotta make a base directory with an `authtok` file,
and a subdirectory for every server you want to connect to.

	BASE_DIRECTORY
	+-- slashnet
	|   +-- handler
	|   +-- config/
	|   |   +-- server
	|   |   +-- gecos
	|   |   +-- nick
	|   +-- log/
	|   |   +-- 2015-01-29T19:56:27Z.log
	|   |   +-- 2015-01-29T20:01:15Z.log
	|   |   +-- 2015-01-29T20:41:40Z.log
	|   |   +-- 2015-01-29T20:41:48Z.log
	|   |   +-- 2015-01-29T20:41:56Z.log
	|   |   +-- 2015-01-29T20:42:44Z.log
	|   +-- outq/
	+-- oftc
	+-- server3
	+-- server4


`config` directory
------------------

The `config` directory in a server directory must have certain files:

* `server` is a list of servers to try and connect to, in the form `hostname:port`
* `gecos` is your "Real Name"
* `nick` is a list of nicknames you'd like to use

The lists are gone through starting with the first entry until one sticks.


`outq` directory
----------------

The `outq` directory is monitored by spongy.
When a new file shows up, its contents are spit out verbatim
over the server connection.

So if you want to send a message to a channel,
do something like this:

	$ echo 'PRIVMSG #channel :hello world' > outq/$$.$(date +%s)


Starting up
-----------

Pretty easy:

	$ cd BASE_DIRECTORY; /path/to/spongy

Spongy will go off and connect to every configured server in BASE_DIRECTORY.


Spongy CGI Configuration
========================

If you'd like to run `spongy.cgi`,
that's fine,
but you have to create a file
called `spongy.basedir`
in the same directory as the CGI.
You can do it like this:

	$ echo '/home/neale/BASE_DIRECTORY' > spongy.basedir

And then,
in `BASE_DIRECTORY`,
you need a file called `auth`
with a sha256 checksum of the authorization token
you want to use in the client.

You can make it like this:

	$ printf 'my fabulous token' | sha256sum | cut -d\  -f1 > BASE_DIRECTORY


Permissions
-----------

There are a lot of different ways to set up permissions.
Here's what I suggest:
make `spongy.cgi` setuid to you.

	$ chmod +s spongy.cgi

If it's setuid,
you don't need to make your config file
(or any other files) 
readable by the user that
runs the web server.

Sadly,
Apache has a whole bunch of weirdness in place
which prevents setuid CGI from working
without a lot of configuration twiddling.
But it also has its own mechanism for running CGI
as the user who owns it.
So if you're using Apache,
please send me a recipe for your solution,
and I'll add it to the distribution :)

