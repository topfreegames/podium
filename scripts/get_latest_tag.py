# -*- coding: utf-8 -*-
# Podium
# https://github.com/topfreegames/podium
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

import urllib
import urllib2
import json

def main():
    url = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:tfgco/podium:pull,push"
    response = urllib.urlopen(url)
    token = json.loads(response.read())['token']

    url = "https://registry-1.docker.io/v2/tfgco/podium/tags/list"
    req = urllib2.Request(url, None, {
        "Authorization": "Bearer %s" % token,
    })
    response = urllib2.urlopen(req)
    tags = json.loads(response.read())
    last_tag = get_last_tag(tags['tags'])
    print last_tag


def get_tag_value(tag):
    if tag[0] == "latest":
        return 0

    while len(tag) < 4:
        tag.append('0')

    total_value = 0
    for index, tag_part in enumerate(tag):
        power = pow(100, len(tag) - index)
        try:
            total_value += int(tag_part) * power
        except:
            total_value = 0
            return

    return total_value


def get_last_tag(tags):
    return '.'.join(
        max([
            (get_tag_value(tag), tag) for tag in
                [t.split('.') for t in tags]
            ], key=lambda i: i[0]
        )[1]
    )


if __name__ == "__main__":
    main()
