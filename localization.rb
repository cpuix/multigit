require 'json'

class Localization
  def initialize
    lang_code = get_system_language
    load_language_file(lang_code)
  end

  def get_message(key)
    keys = key.split('.')
    keys.reduce(@messages) do |acc, k|
      return acc unless acc.is_a?(Hash)
      acc[k]
    end
  end

  private

  def get_system_language
    os = RUBY_PLATFORM
    lang = nil

    if os.include?("darwin")
      lang = IO.popen("defaults read NSGlobalDomain AppleLocale") { |io| io.read.chomp }
    elsif os.include?("linux")
      lang = IO.popen("echo $LANG") { |io| io.read.chomp.split('.').first }
    end

    lang.split('_').first.downcase if lang
  end

  def load_language_file(lang_code)
    file_name = "#{lang_code}.json"
    file_path = File.join(File.dirname(__FILE__), "locales", file_name)
    default_file_path = File.join(File.dirname(__FILE__), "locales", "en.json")
    @messages = File.exist?(file_path) ? JSON.parse(File.read(file_path)) : JSON.parse(File.read(default_file_path))
  end
end