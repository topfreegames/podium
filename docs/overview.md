Overview
========

What is Podium? Podium is a blazing-fast HTTP Leaderboard service. It could be used to manage any number of leaderboards of people or groups, but our aim is players in a game.

Podium allows easy creation of different types of leaderboards with no set-up involved. Create seasonal, localized leaderboards just by varying their names.

## Features

* **Multi-tenant** - Just vary the name of the leaderboard and you can have any number of tenants using leaderboards;
* **Seasonal Leaderboards** - Including suffixes like `year2016week01` or `year2016month06` is all you need to create seasonal leaders. I'm serious! That's all there is to it;
* **No leaderboard configuration** - Just start notifying scores for members of a leaderboard. There's no need to create, configure or maintain leaderboards. Let Podium do that for you;
* **Top Members** - Get the top members of a leaderboard whether you need by absolute value (top 200 members) or percentage (top 3% members);
* **Members around me** - Podium easily returns members around a specific member in the leaderboard. It will even compensate if you ask for the top member or last member to make sure you get a consistent amount of members;
* **Batch score update** - Send a member score to many different leaderboards in a single operation. This allows easy tracking of member rankings in several leaderboards at once (global, regional, clan, etc.);
* **Easy to deploy** - Podium comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

## Architecture

Podium is based on the premise that you have a backend server for your game. That means we only employ basic authentication (if configured).

## The Stack

For the devs out there, our code is in Go, but more specifically:

* Web Framework - [Echo](https://github.com/labstack/echo) based on the insanely fast [FastHTTP](https://github.com/valyala/fasthttp);
* Database - Redis.

## Who's Using it

Well, right now, only us at TFG Co, are using it, but it would be great to get a community around the project. Hope to hear from you guys soon!

## How To Contribute?

Just the usual: Fork, Hack, Pull Request. Rinse and Repeat. Also don't forget to include tests and docs (we are very fond of both).
