Plotomaton
==========

DFA inspired text-based game engine

Creators:
Sean Anderson
Kieron Gillespie
Samuel Payson
Kevin Reid

Maintainer:
Sean Anderson
fnordit@gmail.com

License: GPLv3
    This file is part of Plotomaton.

    Plotomaton is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Plotomaton is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with Plotomaton.  If not, see <http://www.gnu.org/licenses/>.

Project History

Plotomaton started when, in the summer of 2011, I played a bunch of text based adventure games.
That fall, Sam, Kieron, Kevin and I needed a project for a software design and development class,
so we made the system.  It was originally housed on the Clarkson Open Source Institute's servers,
but got deleted at some point, so in fall of 2012 when I started working on it again I put it up
here.

Installation:

Step 1: install go

Step 2: install git

Step 3: set up remote git repository

- Create a directory and go there

- git init

- git remote add origin https://github.com/fnordit/Plotomaton.git

- git pull origin master

Step 4: export GOPATH="/home/[USR]/[DIR]/"

Step 5: modify src/textui/test

Step 6: go run textui.go

It will run the file named "test" stored in the textui directory.
