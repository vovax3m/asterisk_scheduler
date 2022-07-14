# asterisk_scheduler
script makes asterisk call based on prepared csv files

There is a script which reads config files and creates call files in matched time from phone A (phones_from.csv) to phone B (phones_to.scv) at time in <day>.scv or default_schedule.csv of first not found.

# phones_from
`1,101`
index, source number to call from

# phones_to
`1,78001230001,60`
index, dest number, duration

# schedule
`06:00,1,1`
Time for call, index from, index to

#config.conf
configurations for script.

# asterisk
Expected to be configured and ready to make calls with callfiles.

# where to run
Wrap script with service and run on the asterisk node.
