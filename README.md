Spongy IRC
=========

This is a sort of bouncer for clients with transient network connections,
like cell phones and laptops.
It's a lot like [tapchat](https://github.com/tapchat/tapchat) but is a whole lot simpler
while being (at time of writing) much more feature-complete.

It supports (currently) an JavaScript browser-based client,
and can also be worked from the command-line using Unix tools like "tail" and "echo".

Ironically, it doesn't currently work with any existing IRC clients,
although we are kicking around ideas for such a thing.
Honestly, though, if you want a bouncer for a traditional IRC client,
you are better off using something like znc.

We have an [architectural diagram](https://docs.google.com/drawings/d/1am_RTUh89kul-318GoYK73AbjOE_jMYi4vI4NyEgKrY/edit?usp=sharing) if you care about such things.


Features
--------

* Gracefully handles clients with transient networking, such as cell phones and laptops.


Other Documentation
-------------------

* [Installation Instructions](INSTALL.md)
* [Todo list](TODO.md)
