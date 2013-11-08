ghet
====

Personal utility for downloading and updating Github scripts

Install
-------
```sh
$ go install github.com/marksteve/ghet
$ ghet -setup
```


Download
--------
```sh
$ ghet -u https://github.com/jpetazzo/stevedore/blob/master/stevedore -o ~/.bin/stevedore
https://github.com/jpetazzo/stevedore/blob/master/stevedore -> /home/marksteve/.bin/stevedore
```

List
----
```sh
$ ghet -list
Path                           URI
/home/marksteve/.bin/stevedore https://github.com/jpetazzo/stevedore/blob/master/stevedore
```

Update
------
```sh
$ ghet -update -o ~/.bin/stevedore
https://github.com/jpetazzo/stevedore/blob/master/stevedore -> /home/marksteve/.bin/stevedore
```
