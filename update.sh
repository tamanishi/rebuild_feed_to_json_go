#!/bin/bash

cd $(dirname $0)

. .env

./rebuild_feed_to_json_go > episodes.json

git add episodes.json
git commit -m "update episodes"
git push origin master

curl -v -H "Authorization: token ${PRIVATE_ACCESS_TOKEN}" -H "Accept: application/vnd.github.everest-preview+json" "https://api.github.com/repos/tamanishi/rebuildshownotesfilter-nextjs/dispatches" -d '{"event_type": "update-episodes"}'

