#!/usr/bin/env ruby

require "optparse"
require "hashdiff"
require "yaml"
require "pp"
require "deepsort"

options = {
	:write_sorted_files => false,
	:include_expected_differences => false,
}

OptionParser.new do |opts|
  opts.banner = "Usage: yaml-comparer.rb [options] OLDFILE NEWFILE"

  opts.on("--write-sorted-files", "Write sorted files") do |w|
    options[:write_sorted_files] = w
  end

  opts.on("--include-expected-differences", "Include expected differences in the diff result") do |i|
    options[:include_expected_differences] = i
  end
end.parse!

if ARGV.size != 2
	puts "Incomplete parameter list. Exiting..."
  exit 1
end

left_file = ARGV.shift
right_file = ARGV.shift

oldyaml = YAML::load(File.open(left_file))
newyaml = YAML::load(File.open(right_file))

oldyaml.deep_sort!
newyaml.deep_sort!

#Sort the yaml contents by kind and then by name on the objects array. We do this because otherwise
#would be sorted by APIVersion and that would give us lots of order changes at that level,
#making the diff unreadable
oldyaml["objects"].sort_by! { |obj| obj["kind"].downcase+obj["metadata"]["name"].downcase} #TODO how to sort first by kind and then by name without hacks?
newyaml["objects"].sort_by! { |obj| obj["kind"].downcase+obj["metadata"]["name"].downcase} #TODO how to sort first by kind and then by name without hacks?

# Set this to true if you want the deep sort results to be written to
# the new files

if options[:write_sorted_files]
  File.open("oldsortedyaml.yml", "w") { |file| file.write(oldyaml.to_yaml) }
	File.open("newsortedyaml.yml", "w") { |file| file.write(newyaml.to_yaml) }
end

if oldyaml == newyaml
  puts "YAML files are equal"
else
  puts "YAML files are different"
end

puts "Differences:"
diff = HashDiff.diff(oldyaml, newyaml, {similarity: 0.0})

# We ignore some of the diff results that we expect will be shown
# but are expected differences (for example, fields that are automatically added by the
# Kubernetes API deserializer when they not exist in the original file, like the
# creationTimestamp field
element_types_regexes=[
 '^base_env$',
 '^metadata\.creationTimestamp$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.to\.weight$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.template\.spec\.initContainers\[\d+\]\.resources$',  #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.strategy\.resources$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.containers\[\d+\]\.resources\.requests', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.requests$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.requests\.memory$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.requests\.cpu$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.limits$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.limits\.memory$',#we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.spec\.template\.spec\.containers\[\d+\]\.resources\.limits\.cpu$', #we manually removed it from the parser due parsing problems
 '^objects\[\d+\]\.metadata\.creationTimestamp$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.template\.metadata\.creationTimestamp$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.status$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.test$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.lookupPolicy$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.test$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.tags\[\d+\]\.generation$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.tags\[\d+\]\.referencePolicy$', #automatically added by the go generator when does not exist in the original yaml
 '^objects\[\d+\]\.spec\.dataSource$', #automatically added by the go generator when does not exist in the original yaml

 #'^objects\[\d+\]$', # TODO is correct to use this regex? even though a fully new object is added would the individual diffs be shown before??
 #'^parameters\[\d+\].*$', # TODO remove this when you want to see the parameters differences
]

if !options[:include_expected_differences]
  element_types_regexes.each do |regex_elem|
		diff.reject! { |item| item[1] =~ /#{regex_elem}/ }

		#Special cases:
		# When the annotations field exists in the original file but is empty then the new generator removes it
		diff.reject! { |item| item[0] == "-" && item[1] =~ /'^objects\[\d+\]\.metadata\.annotations$'/ && item[2] == nil }
		# When no strategy resources are specified in the original file then an empty resources field is added by the go generator
		diff.reject! { |item| item[0] == "+" && item[1] =~ /'^objects\[\d+\]\.spec\.strategy$'/ && item[2] == {"resources"=>{}} }

		# Also, on zync-database-data volume spec some expected zync changes will be visible but no check has
		# been implemented for it yet
	end
end

pp diff

exit 0
