require_relative 'localization'
require_relative 'validation'

module InputManager
  def self.get_valid_input(message_key, validation_method)
    loop do
      puts Localization.get_message(message_key)
      input = STDIN.gets.chomp

      return input if send(validation_method, input)
      return nil if input.downcase == 'exit' || input.downcase == 'quit'

      puts Localization.get_message("error.#{validation_method}")
    end
  end

  private

  def self.valid_account_name?(name)
    Validation.valid_account_name?(name)
  end

  def self.valid_email?(email)
    Validation.valid_email?(email)
  end
end
