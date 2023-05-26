defmodule ElkiaTest do
  use ExUnit.Case
  doctest Elkia

  test "greets the world" do
    assert Elkia.hello() == :world
  end
end
