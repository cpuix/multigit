require 'test/unit'
require_relative 'multigit'

class TestMultiGit < Test::Unit::TestCase
  def setup
    @multigit_create = MultiGit.new('create', 'test_account', 'test@example.com')
    @multigit_delete = MultiGit.new('delete', 'test_account')
    @multigit_copy = MultiGit.new('copy', 'test_account')
    def @multigit_copy.exit(code=0)
      raise "Exit called with code #{code}"
    end
  end

  def test_create_account_key
    assert_nothing_raised do
      @multigit_create.execute
    end
  end

  def test_delete_account_key
    assert_nothing_raised do
      @multigit_delete.execute
    end
  end

  def test_copy_public_key
    assert_raise(RuntimeError) do
      @multigit_copy.execute
    end
  end
end