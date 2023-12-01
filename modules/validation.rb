# frozen_string_literal: true

module Validation
  def self.valid_account_name?(account_name)
    account_name.match?(/^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$/i)
  end

  def self.valid_email?(email)
    email.match?(/\A[\w+\-.]+@[a-z\d\-]+(\.[a-z\d\-]+)*\.[a-z]+\z/i)
  end
end
