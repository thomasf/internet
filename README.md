# internet :) [![Build Status](https://drone.io/github.com/thomasf/internet/status.png)](https://drone.io/github.com/thomasf/internet/latest)


This package contains a server and client for historical searches in bgp dumps. I'm not sure where this is going so for the moment this package is named internet to postpone the definition.

The first two things to be implemented are ports of the core functionality in  https://github.com/CIRCL/IP-ASN-history and https://github.com/CIRCL/ASN-Description-History .


## pre requirements

Download and install bgpdump: http://www.ris.ripe.net/source/bgpdump/


## whats done

* download bgpdump by date
* import downloaded bgpdump files
* query full ip2asn history
* add query for latest imported date only

## todo (shortlist)
* add ASN2Description support.
* add dump update scheduling function suitable to be integrated into a query server
* replace cmd/dev with an example
* write tests
* final API design
* improve docs

# maybe/later

* add query for specific date lookup
* support removing old dates (could it be better to just rebuild from a separate redis db?)



