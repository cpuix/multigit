require 'tempfile'
require 'open3'

class KeyManager
  def initialize(config, localization)
    @config = config
    @localization = localization
  end

  def self.create_ssh_key(account_name,account_email, add_passphrase = false)
    command = "ssh-keygen -t ed25519 -C '#{account_email}' -f #{self.ssh_key_path(account_name)}"
    if add_passphrase
      command += " -N ''"
    end
    Open3.popen3(command) do |stdin, stdout, stderr, wait_thr|
      error_message = stderr.read
      if error_message.empty?
        puts "SSH key successfully generated."
        return true
      else
        warn "Error generating SSH key: #{error_message}"
        return false
      end
    end
  end
  def self.add_ssh_config_entry(account_name)
    key_path = self.ssh_key_path(account_name)
    begin
      ssh_config_content = File.read(SSH_CONFIG_PATH)
    rescue => e
      warn "SSH konfigürasyon dosyası okunamadı: #{e.message}"
      return false
    end

    config_entry = <<~CONFIG
      Host github.com-#{account_name}
      HostName github.com
      User git
      IdentityFile #{key_path}
    CONFIG

    # Check current input
    existing_entry_regex = /Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{key_path}/
    existing_entry = ssh_config_content.match(existing_entry_regex)

    # New input
    updated_content = if existing_entry
                        ssh_config_content.gsub(existing_entry.to_s, config_entry)
                      else
                        ssh_config_content + "\n#{config_entry}"
                      end

    # Use Tempfile for Atomic update
    Tempfile.create('ssh_config') do |tempfile|
      begin
        tempfile.write(updated_content)
        tempfile.close
        FileUtils.mv(tempfile.path, SSH_CONFIG_PATH)
      rescue => e
        warn "SSH konfigürasyon dosyası güncellenirken hata oluştu: #{e.message}"
        return false
      end
    end

    true
  end

  def self.remove_ssh_config_entry(account_name)
    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")
    File.write(SSH_CONFIG_PATH, File.read(SSH_CONFIG_PATH).gsub(/Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{ssh_key_file}\n/, ''))
  end

  def self.delete(account_name)
    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")
    FileUtils.rm_rf([ssh_key_file, "#{ssh_key_file}.pub"])
  end

  def self.check_key_exists(account_name)
    if File.exist?(self.ssh_key_path(account_name))
      puts Localization.get_message("ssh.key_exists")
      exit 1
    end
    true
  end

  def self.add_key_to_agent(account_name)
    command = "ssh-add --apple-use-keychain #{self.ssh_key_path(account_name)}"

    Open3.popen3(command) do |stdin, stdout, stderr, wait_thr|
      error_message = stderr.read
      if error_message.empty?
        puts "SSH key added to agent successfully."
        return true
      else
        warn "Error adding SSH key to agent: #{error_message}"
        return false
      end
    end
  end

  def self.ssh_key_path(account_name)
    File.join(ENV['HOME'], '.ssh', "github-#{account_name}")
  end

  def self.copy_public_key_to_clipboard(account_name)
    public_key_file = "#{self.ssh_key_path(account_name)}.pub"

    begin
      public_key_content = File.read(public_key_file)
    rescue => e
      warn "SSH anahtar dosyası okunamadı: #{e.message}"
      return false
    end

    begin
      IO.popen('pbcopy', 'w') { |f| f << public_key_content }
      puts "SSH public key copied to clipboard."
    rescue => e
      warn "SSH anahtarını panoya kopyalarken hata oluştu: #{e.message}"
      return false
    end

    true
  end

end
