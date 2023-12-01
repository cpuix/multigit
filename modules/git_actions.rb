# frozen_string_literal: true

module GitActions

  def self.getAccount(account_name)
    account = Account.find_by(name: account_name)
    if account.nil?
      raise ArgumentError, @localization.get_message("system.account_not_found")
    end
    account
  end
end
