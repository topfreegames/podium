#!/bin/bash

if [ "$BASICAUTH_USERNAME" == "" ]; then
  cp /home/podium/nginx_default /etc/nginx/sites-enabled/default
else
  echo configuring htpasswd
  htpasswd -b -c /etc/nginx/.htpasswd $BASICAUTH_USERNAME $BASICAUTH_PASSWORD
  cp /home/podium/nginx_default_basicauth /etc/nginx/sites-enabled/default
fi

echo configuring nginx
sed -i "s/{SERVER_NAME}/$SERVER_NAME/g" /etc/nginx/sites-enabled/default

echo starting podium
/usr/bin/supervisord --nodaemon --configuration /etc/supervisord.conf
