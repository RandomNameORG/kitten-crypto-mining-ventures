#!/usr/bin/env ruby
# Rewrites Formula/meowmine.rb to match the given VERSION and the sha256s of the
# built binaries in bin/. Used by the release workflow after a tag build so the
# Homebrew formula always tracks the latest release instead of a pinned version.

require "digest"

version = ARGV[0]
abort "usage: update-formula.rb <version>" if version.nil? || version.empty?
version = version.sub(/\Av/, "")

formula_path = File.expand_path("../Formula/meowmine.rb", __dir__)
bin_dir = File.expand_path("../bin", __dir__)

binaries = %w[
  meowmine-darwin-arm64
  meowmine-ssh-darwin-arm64
  meowmine-darwin-amd64
  meowmine-ssh-darwin-amd64
  meowmine-linux-amd64
  meowmine-ssh-linux-amd64
]

shas = binaries.to_h do |name|
  path = File.join(bin_dir, name)
  abort "missing binary: #{path}" unless File.exist?(path)
  [name, Digest::SHA256.file(path).hexdigest]
end

content = File.read(formula_path)
content.sub!(/^(\s*version )"[^"]*"/, "\\1\"#{version}\"")

binaries.each do |name|
  pattern = /(url "\#\{base\}\/#{Regexp.escape(name)}"\s*\n\s*sha256 ")[0-9a-f]{64}(")/
  unless content.sub!(pattern, "\\1#{shas[name]}\\2")
    abort "could not find sha256 line for #{name} in formula"
  end
end

File.write(formula_path, content)
puts "Updated #{formula_path} to version #{version}"
