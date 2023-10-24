#!/usr/bin/env ruby

require 'fileutils'

SSH_CONFIG_PATH = File.join(ENV['HOME'],'.ssh','config')

class MultiGit
  def initialize(command, *args)
    @command = command
    @args = args
  end

  def execute
    case @command
    when 'create'
      create_account_key(*@args)
    when 'delete'
      delete_account_key(*@args)
    when 'copy'
      copy_public_key(*@args)
    else
      puts "Incorrect command. Available commands: create, delete, copy"
      exit 1
    end
  end

  private

  def create_account_key(account_name, user_email)
    validate_account_name(account_name)
    validate_user_email(user_email)

    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    if File.exist?(ssh_key_file)
      puts "A key file with this name already exists."
      exit 1
    end

    puts "Would you like to add a password? (y/n)"
    add_passphrase = STDIN.gets.chomp

    if add_passphrase == 'y'
      system("ssh-keygen -t ed25519 -C '#{user_email}' -f #{ssh_key_file}")
    else
      system("ssh-keygen -t ed25519 -C '#{user_email}' -f #{ssh_key_file} -N ''")
    end

    system("ssh-add -K #{ssh_key_file}")

    config_entry = <<~CONFIG
      Host github.com-#{account_name}
      HostName github.com
      User git
      IdentityFile #{ssh_key_file}
    CONFIG

    existing_entry = File.read(SSH_CONFIG_PATH).match(/Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{ssh_key_file}/)

    if existing_entry.nil?
      File.open(SSH_CONFIG_PATH, 'a') { |f| f.write("\n#{config_entry}") }
    else
      File.write(SSH_CONFIG_PATH, File.read(SSH_CONFIG_PATH).gsub(existing_entry.to_s, config_entry))
    end

    puts "SSH keys (pub and sub) have been created and added to the config file. Don't forget to manually add the keys to your GitHub account.\n"
    puts "Public Key content you need to copy to add to GitHub:"
    puts File.read("#{ssh_key_file}.pub")
  end

  def delete_account_key(account_name)
    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    unless File.exist?(ssh_key_file)
      puts "There is no key file with that name."
      exit 1
    end

    puts "This operation cannot be reversed. Do you want to continue? (y/n)"
    confirm_delete = STDIN.gets.chomp

    if confirm_delete == 'y'
      FileUtils.rm_rf([ssh_key_file, "#{ssh_key_file}.pub"])
      File.write(SSH_CONFIG_PATH, File.read(SSH_CONFIG_PATH).gsub(/Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{ssh_key_file}\n/, ''))
      puts "The key file and the entry in the config file have been deleted."
    else
      puts "The transaction has been cancelled."
    end
  end

  def copy_public_key(account_name)
    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    unless File.exist?(ssh_key_file)
      puts "There is no key file with that name."
      exit 1
    end

    IO.popen('pbcopy', 'w') { |f| f << File.read("#{ssh_key_file}.pub") }
    puts "github-#{account_name} Public Key copied."
  end

  def validate_account_name(account_name)
    unless account_name =~ /^[a-zA-Z0-9_-]+$/
      puts "Invalid characters were used. Only letters, numbers, underscores and hyphens can be used."
      exit 1
    end
  end

  def validate_user_email(user_email)
    unless user_email =~ /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
      puts "Invalid email address."
      exit 1
    end
  end
end

if __FILE__ == $0
  MultiGit.new(*ARGV).execute
end
