#!/bin/sh
while true; do
  ZID=$(awslocal route53 list-hosted-zones --query "HostedZones[0].Id" --output text 2>/dev/null | sed "s#/hostedzone/##")
  {
    echo '$ORIGIN test.'
    echo '$TTL 60'
    echo "@ IN SOA ns.test. admin.test. $(date +%s) 7200 3600 1209600 60"
    echo '@ IN NS ns.test.'
    echo 'ns IN A 127.0.0.1'
    [ -n "$ZID" ] && awslocal route53 list-resource-record-sets --hosted-zone-id "$ZID" --output json 2>/dev/null | python3 -c '
import json,sys
try: rs=json.load(sys.stdin).get("ResourceRecordSets",[])
except Exception: rs=[]
for r in rs:
    if r.get("Type")!="TXT": continue
    for v in r.get("ResourceRecords",[]):
        print(r["Name"], "IN TXT", v["Value"])
'
  } > /zones/test.zone.tmp && mv /zones/test.zone.tmp /zones/test.zone
  sleep 2
done
