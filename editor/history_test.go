package editor

import (
	"testing"
	"time"
)

// fakeClock lets tests decide whether edits are merged into one step
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

func historyTestApp(t *testing.T) (*App, *fakeClock) {
	app := paintTestApp(t)
	clock := &fakeClock{t: time.Unix(1000, 0)}
	app.history.now = clock.now
	return app, clock
}

func renamePlayer(t *testing.T, app *App, index int, name string) {
	players := cloneValue(app.GetPlayers())
	players[index].LeaderName = name
	if err := app.SetPlayers(players); err != nil {
		t.Fatal(err)
	}
}

func TestUndoPlayersMergesQuickEdits(t *testing.T) {
	app, clock := historyTestApp(t)
	original := app.wbMap.Players[0].LeaderName

	// Typing a name: every key press is a separate SetPlayers call
	for _, name := range []string{"C", "Ca", "Cae", "Caesar"} {
		renamePlayer(t, app, 0, name)
		clock.t = clock.t.Add(200 * time.Millisecond)
	}
	if s := app.HistoryState(); s.Undo != "players" {
		t.Fatalf("undo label = %q", s.Undo)
	}
	clock.t = clock.t.Add(5 * time.Second)
	renamePlayer(t, app, 1, "Second")

	app.Undo()
	if got := app.wbMap.Players[1].LeaderName; got == "Second" {
		t.Error("the later edit must be a separate step")
	}
	if got := app.wbMap.Players[0].LeaderName; got != "Caesar" {
		t.Errorf("the first edit must stay, got %q", got)
	}
	app.Undo()
	if got := app.wbMap.Players[0].LeaderName; got != original {
		t.Errorf("typing must be undone at once, got %q", got)
	}
	if s := app.HistoryState(); s.Undo != "" || s.Redo != "players" {
		t.Errorf("state after undo = %+v", s)
	}

	app.Redo()
	if got := app.wbMap.Players[0].LeaderName; got != "Caesar" {
		t.Errorf("redo must restore the whole name, got %q", got)
	}
}

func TestEditAfterUndoIsNotMerged(t *testing.T) {
	app, _ := historyTestApp(t)
	renamePlayer(t, app, 0, "A")
	renamePlayer(t, app, 0, "B")
	app.Undo()
	// The same moment, but the undone step must not swallow the new edit
	renamePlayer(t, app, 0, "C")
	if n := len(app.history.undo); n != 1 {
		t.Fatalf("expected one step, got %d", n)
	}
	app.Undo()
	if got := app.wbMap.Players[0].LeaderName; got == "B" || got == "C" {
		t.Errorf("undo must return to the state before the edits, got %q", got)
	}
}

func TestUndoSectionsAndUnchangedValues(t *testing.T) {
	app, clock := historyTestApp(t)

	game := cloneValue(app.GetGame())
	game.Era = "ERA_MEDIEVAL"
	if err := app.SetGame(game); err != nil {
		t.Fatal(err)
	}
	// Changing the stored value in place must not change the history
	app.wbMap.Game.Era = "ERA_MODERN"
	clock.t = clock.t.Add(time.Minute)

	teams := cloneValue(app.GetTeams())
	teams[0].Tech = append(teams[0].Tech, "TECH_MINING")
	if err := app.SetTeams(teams); err != nil {
		t.Fatal(err)
	}
	clock.t = clock.t.Add(time.Minute)

	props := cloneValue(app.GetMapProps())
	props.WrapY = 1 - props.WrapY
	if err := app.SetMapProps(props); err != nil {
		t.Fatal(err)
	}
	clock.t = clock.t.Add(time.Minute)

	// Saving the same values (e.g. when an editor reloads) is not a step
	if err := app.SetPlayers(cloneValue(app.GetPlayers())); err != nil {
		t.Fatal(err)
	}
	if n := len(app.history.undo); n != 3 {
		t.Fatalf("expected 3 steps, got %d", n)
	}

	app.Undo()
	if app.wbMap.Map.WrapY == props.WrapY {
		t.Error("map props must be restored")
	}
	app.Undo()
	if len(app.wbMap.Teams[0].Tech) != len(teams[0].Tech)-1 {
		t.Errorf("techs must be restored: %v", app.wbMap.Teams[0].Tech)
	}
	app.Undo()
	if app.wbMap.Game.Era == "ERA_MEDIEVAL" {
		t.Error("game must be restored")
	}
	app.Redo()
	if app.wbMap.Game.Era != "ERA_MEDIEVAL" {
		t.Errorf("redo must restore the stored copy, got %q", app.wbMap.Game.Era)
	}
}

func TestUndoSwapAndClearPlayer(t *testing.T) {
	app, _ := historyTestApp(t)
	m := app.wbMap
	m.Players[0].LeaderName, m.Players[1].LeaderName = "First", "Second"
	plot := app.GetPlot(2, 2)
	plot.Units = []*Unit{{UnitType: "UNIT_WARRIOR", UnitOwner: 0}}
	plot.Cities = []*City{{CityName: "Rome", CityOwner: 0}}

	if err := app.SwapPlayers(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := app.ClearPlayer(0, true); err != nil {
		t.Fatal(err)
	}
	if s := app.HistoryState(); s.Undo != "clear:0" {
		t.Errorf("undo label = %q", s.Undo)
	}
	if got := app.GetPlot(2, 2); got.Units[0].UnitOwner != 1 || got.Cities[0].CityOwner != 1 {
		t.Fatalf("assets of the swapped player must stay: %+v", got)
	}

	app.Undo()
	if app.wbMap.Players[0].LeaderName != "Second" {
		t.Errorf("clear must be undone: %+v", app.wbMap.Players[0])
	}
	app.Undo()
	got := app.GetPlot(2, 2)
	if app.wbMap.Players[0].LeaderName != "First" || got.Units[0].UnitOwner != 0 || got.Cities[0].CityOwner != 0 {
		t.Errorf("swap must be undone: %+v %+v", app.wbMap.Players[0], got)
	}
}
