#!/bin/bash

rm -r build/*
mkdir build/

~/go/bin/fyne package -icon resources/MidiOn.png --appVersion 2.0.0
tar xf midi2key-ng.tar.xz -C build/
mv midi2key-ng.tar.xz build/usr/local/bin/midi2key-ng-linux-amd64.tar.xz
mv build/usr/local/bin/midi2key-ng build/usr/local/bin/midi2key-ng-linux-amd64
~/go/bin/selfupdatectl sign build/usr/local/bin/midi2key-ng-linux-amd64

mv build/usr/local/bin/* build/
rm build/Makefile
rm -r build/usr