#!/usr/bin/env python3

import boto3
import datetime
import glob
import json
import multiprocessing
import operator
import os
import pytz
import sys
import time
import yaml
import math

from minio import Minio
from minio.error import ResponseError


dtfmt = "%Y-%m-%d %H:%M %Z"
utc = pytz.timezone('UTC')
pst = pytz.timezone('US/Pacific')


if len(sys.argv) < 3:
	print("error: missing required command line parameter")
	print("usage: {} <input-file> <output-file>".format(sys.argv[0]))
	sys.exit(1)

input_file = sys.argv[1]
output_file = sys.argv[2]

print("Reading from {}".format(input_file))
print("Writing to {}".format(output_file))

ds = json.loads(open(input_file).read())

def chunkify(num_total, num_parts):
	nper = math.floor(num_total / num_parts)
	if num_total % num_parts != 0:
		nper += 1
	return [[i*nper, i*nper+nper] if i*nper < num_total else [i*nper, num_total] for i in range(num_parts)]

user = os.environ['AWS_ACCESS_KEY_ID']
pw = os.environ['AWS_SECRET_ACCESS_KEY']
s3_ep = os.environ['S3_ENDPOINT_URL']
s3_ep_sp = s3_ep.split('://')
host = s3_ep_sp[1] if len(s3_ep_sp) > 1 else s3_ep_sp[0]
bucket = os.environ['S3_BUCKET_ID']
prefix = 'build_cache/'

def initialize():
	global mc
	mc = Minio(host, access_key=user, secret_key=pw, secure=True)

def stat(obj):
	p = "{}{}".format('build_cache/', obj['specfile'])
	try:
		print("stat'ing {}".format(p))
		so = mc.stat_object(bucket, p)
		dt = datetime.datetime(*so.last_modified[:6], tzinfo=utc)
		dt = dt.astimezone(pst)
		obj["last_modified_pretty"] = dt.strftime(dtfmt)
		obj["last_modified"] = time.mktime(dt.timetuple())
		obj["versioned_name"] = "{}@{}".format(obj["name"], obj["version"])
		return obj
	except:
		print("FAILED STAT: {}".format(p))
		return None

objects = [v for _, v in ds.items()]
results = []

t1 = time.time()

cpu_count = multiprocessing.cpu_count()
if cpu_count > 128:
	cpu_count = 128
print("Using CPU Count: {}".format(cpu_count))
pool = multiprocessing.Pool(cpu_count, initialize)
results = pool.map(stat, objects)
pool.close()
pool.join()

lastmod = -1
lastmod_pretty = ""
for r in results:
	if r is None:
		continue
	if r['last_modified'] > lastmod:
		lastmod = r['last_modified']
		lastmod_pretty = r['last_modified_pretty']

print("Last Modified: {}".format(lastmod))
print("Last Modified Pretty: {}".format(lastmod_pretty))
print("# results = {}".format(len(results)))
print("{} seconds".format(time.time() - t1))

specs = {}
for r in results:
	k = r['versioned_name']
	if k not in specs:
		specs[k] = []
	specs[k].append(r)

for k, v in specs.items():
  v.sort(key=operator.itemgetter('arch', 'os', 'compiler', 'last_modified'))

fjs = {
	"meta": {
		"last_mod": lastmod_pretty
	},
	"data": []
}

ns = list(specs.keys())
ns.sort()

for n in ns:
  o = { "name": specs[n][0]["name"], "objs": specs[n] }
  fjs["data"].append(o)

with open(output_file, 'w') as f:
  f.write(json.dumps(fjs, indent=1))
