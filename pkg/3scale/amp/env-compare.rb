require "hashdiff"
require "json"
require "deepsort"

String.module_exec do
  def nil?
    empty?
  end
end

envs = ARGV.map { |f| JSON.parse(File.read(f)).deep_sort }
diff = HashDiff.diff(*envs, similarity: 0.0)

diff.each do |change| change
  sign, key, *values = *change
  puts "#{sign} #{key} #{values.map(&:inspect).join(' <=> ')}"
end