import json

ds = json.loads(open('inventory.json').read())['data']

data = []
for d in ds:
    data += d['objs']

archVals = set([d['arch'] for d in data])
osVals = set([d['os'] for d in data])

archValCounts = {}
osValCounts = {}
for d in data:
    archval = d['arch']
    if archval not in archValCounts:
        archValCounts[archval] = 0
    archValCounts[archval] += 1

    osval = d['os']
    if osval not in osValCounts:
        osValCounts[osval] = 0
    osValCounts[osval] += 1

threshold = 1000
print([v for v in archVals if archValCounts[v] > threshold])
print([v for v in osVals if osValCounts[v] > threshold])
