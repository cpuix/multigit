class Validation
  def self.valid_account_name?(account_name)
    /^[a-zA-Z0-9_-]+$/.match?(account_name)
  end

  def self.valid_email?(email)
    /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.match?(email)
  end
end