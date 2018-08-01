#!/bin/bash

notify_mail() {
	aws sns publish --topic-arn $1 --message "$2" --region $3 > /dev/null 2>&1
}

get() {
	aws s3 sync --quiet $1 $2 --delete
	r=$?
	if [ ${r} != 0 ]; then
		notify_mail $3 "Failed to download the content file. Status code: $r" $4
	fi
	return $r
}

BUCKET=s3://bucket-name
TARGETDIR=/var/www/html/
ARN=SNS-topic-ARN
REGION=region

get $BUCKET $TARGETDIR $ARN $REGION
