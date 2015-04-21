# internet :) [![Build Status](https://drone.io/github.com/thomasf/internet/status.png)](https://drone.io/github.com/thomasf/internet/latest)

This package contains a server and client for historical searches in bgp dumps.
I'm not sure where this is going so for the moment this package is named
internet to postpone the definition.

## Features

* Downloads BGP table dumps from http://data.ris.ripe.net/rrc00 and imports
  them into redis for current and historical IP address to AS Number lookup.
* Downloads http://www.cidr-report.org/as2.0/autnums.html (controlled to once
  per day) and stores the entries in redis for current and historical AS Number
  to AS Description lookup.
* Caches all data downloads so that databases can be rebuilt easily.


## Pre requirements

* BGPDump - [Download](http://www.ris.ripe.net/source/bgpdump/), compile and
  install it somewhere into PATH.
* Redis


## Acknowledgments

Basic design for the IP2ASN history and ASN2ASDescription parts were inspired
from https://github.com/CIRCL/IP-ASN-history and
https://github.com/CIRCL/ASN-Description-History .
