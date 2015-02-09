wirc protocol
===========

This document attempts to describe the wirc protocol.
The source code will always be authoritative,
but this protocol has been around for a while now so should be changed often, if ever.


Out Queue
---------


Any files that appear in the directory outq/ are written verbatim to the IRC server.

You may put multiple lines in a single file.

Filenames beginning with "." are ignored.

You are advised to create files beginning with ".",
then rename them on completion of the write,
to avoid race conditions.


Log
---

### Log Filenames

TBD


### Log Messages

IRC messages are written to the log, one message per line.
Messages are translated to an easier-to-parse format:

	timestamp fullname command sender forum [args...] :text
	
Where:

* `timestamp` is in Unix epoch time.
* `fullname` is the full name of the message origin (typically `nick!user@host.name`)
* `command` is the IRC command
* `sender` is the IRC name of the entity that sent the message
* `forum` is the IRC name of the audience of the message
* `args` are any additional arguments not otherwise covered
* `text` is the text of the message

`sender` and `forum` are provided in every message, for the convenience of the client.
A PRIVMSG to `sender` will make it back to whomever sent the message,
a PRIVMSG to `forum` will be seen by everyone in the audience.

For example, a "private message" will have `sender` equal to `forum`.
But a "channel message" will have `forum` set to the channel.

See `wirc.go` for details of each message type.


### Initial Messages

Each log file will contain the following initial messages,
to facilitate stateful clients:

TBD
