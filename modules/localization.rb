require 'json'

module Localization
  @messages = {}
  @default_locale = 'en'

  def self.initialize
    lang_code = get_system_language || @default_locale
    load_language_file(lang_code)
  end

  def self.get_message(key)
    keys = key.split('.')
    keys.reduce(@messages) do |acc, k|
      return acc unless acc.is_a?(Hash)
      acc[k]
    end
  end

  private_class_method

  def self.get_system_language
    os = RUBY_PLATFORM
    lang = nil

    if os.include?("darwin")
      lang = IO.popen("defaults read NSGlobalDomain AppleLocale") { |io| io.read.chomp }
    elsif os.include?("linux")
      lang = IO.popen("echo $LANG") { |io| io.read.chomp.split('.').first }
    end

    lang.split('_').first.downcase if lang
  end

  def self.load_language_file(lang_code)
    file_name = "#{lang_code}.json"
    file_path = File.join(File.expand_path('../../', File.expand_path(__FILE__)), "locales", file_name)
    default_file_path = File.join(File.expand_path('../../', File.expand_path(__FILE__)), "locales", "en.json")
    file_to_load = File.exist?(file_path) ? file_path : default_file_path
    @messages = JSON.parse(File.read(file_to_load))
  end
end

Localization.initialize