import sys, json
d = json.load(sys.stdin)
term = d['status'] not in ('running', 'pending')
print(('TERM ' if term else 'BUSY ') + d['status'])
for s in d.get('stages', []):
    bad = [st['name'] for st in s.get('steps', []) if st['status'] not in ('success', 'skipped')]
    run = [st['name'] for st in s.get('steps', []) if st['status'] == 'running']
    print('  %-28s %-9s %s' % (s['name'], s['status'],
          ('FAIL:' + ','.join(bad)) if (term and bad) else ('at:' + run[0] if run else '')))
