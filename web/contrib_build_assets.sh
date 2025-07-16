#!/bin/bash

SELF=$(cd `dirname $0`; pwd)
cd $SELF
mkdir -p assets
cd assets
curl -L -o bootstrap-3.3.7.min.css https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css
curl -L -o bootstrap-3.3.7.min.js https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js
curl -L -o jquery-1.12.4.min.js https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js
