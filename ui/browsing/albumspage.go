package browsing

import (
	"supersonic/backend"
	"supersonic/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*AlbumsPage)(nil)

type AlbumsPage struct {
	widget.BaseWidget

	title       string
	im          *backend.ImageManager
	lm          *backend.LibraryManager
	nav         func(Route)
	grid        *widgets.AlbumGrid
	searchGrid  *widgets.AlbumGrid
	searcher    *widgets.Searcher
	searchText  string
	titleDisp   *widget.RichText
	sortOrder   *selectWidget
	container   *fyne.Container
	OnPlayAlbum func(string, int)
}

type selectWidget struct {
	widget.Select
	height float32
}

func NewSelect(options []string, onChanged func(string)) *selectWidget {
	s := &selectWidget{
		Select: widget.Select{
			Options:   options,
			OnChanged: onChanged,
		},
	}
	s.height = widget.NewSelect(nil, nil).MinSize().Height
	s.ExtendBaseWidget(s)
	return s
}

func (s *selectWidget) MinSize() fyne.Size {
	return fyne.NewSize(170, s.height)
}

func NewAlbumsPage(title string, sortOrder string, lm *backend.LibraryManager, im *backend.ImageManager, nav func(Route)) *AlbumsPage {
	a := &AlbumsPage{
		title: title,
		lm:    lm,
		im:    im,
		nav:   nav,
	}
	a.ExtendBaseWidget(a)

	a.titleDisp = widget.NewRichTextWithText(title)
	a.titleDisp.Segments[0].(*widget.TextSegment).Style = widget.RichTextStyle{
		SizeName: theme.SizeNameHeadingText,
	}
	a.sortOrder = NewSelect(backend.AlbumSortOrders, nil)
	a.sortOrder.Selected = sortOrder
	a.sortOrder.OnChanged = a.onSortOrderChanged
	iter := lm.AlbumsIter(backend.AlbumSortOrder(a.sortOrder.Selected))
	a.grid = widgets.NewAlbumGrid(iter, im, false /*showYear*/)
	a.grid.OnPlayAlbum = a.onPlayAlbum
	a.grid.OnShowArtistPage = a.onShowArtistPage
	a.grid.OnShowAlbumPage = a.onShowAlbumPage
	a.searcher = widgets.NewSearcher()
	a.searcher.OnSearched = a.OnSearched
	a.createContainer()

	return a
}

func (a *AlbumsPage) createContainer() {
	searchVbox := container.NewVBox(layout.NewSpacer(), a.searcher.Entry, layout.NewSpacer())
	sortVbox := container.NewVBox(layout.NewSpacer(), a.sortOrder, layout.NewSpacer())
	a.container = container.NewBorder(
		container.NewHBox(widgets.NewHSpace(9), a.titleDisp, sortVbox, layout.NewSpacer(), searchVbox, widgets.NewHSpace(15)),
		nil,
		nil,
		nil,
		a.grid,
	)
}

func restoreAlbumsPage(saved *savedAlbumsPage) *AlbumsPage {
	a := &AlbumsPage{
		title: saved.title,
		lm:    saved.lm,
		im:    saved.im,
		nav:   saved.nav,
	}
	a.ExtendBaseWidget(a)

	a.titleDisp = widget.NewRichTextWithText(a.title)
	a.titleDisp.Segments[0].(*widget.TextSegment).Style = widget.RichTextStyle{
		SizeName: theme.SizeNameHeadingText,
	}
	a.sortOrder = NewSelect(backend.AlbumSortOrders, nil)
	a.sortOrder.Selected = saved.sortOrder
	a.sortOrder.OnChanged = a.onSortOrderChanged
	a.grid = widgets.NewAlbumGridFromState(saved.gridState)
	a.searcher = widgets.NewSearcher()
	a.searcher.OnSearched = a.OnSearched
	a.createContainer()

	return a
}

func (a *AlbumsPage) OnSearched(query string) {
	a.searchText = query
	if query == "" {
		a.container.Objects[0] = a.grid
		if a.searchGrid != nil {
			a.searchGrid.Clear()
		}
		a.Refresh()
		return
	}
	a.doSearch(query)
}

func (a *AlbumsPage) Route() Route {
	return AlbumsRoute(backend.AlbumSortOrder(a.sortOrder.Selected))
}

var _ Searchable = (*AlbumsPage)(nil)

func (a *AlbumsPage) SearchWidget() fyne.Focusable {
	return a.searcher.Entry
}

func (a *AlbumsPage) SetPlayAlbumCallback(cb func(string, int)) {
	a.OnPlayAlbum = cb
}

func (a *AlbumsPage) Reload() {
	if a.searchText != "" {
		a.doSearch(a.searchText)
	} else {
		a.grid.Reset(a.lm.AlbumsIter(backend.AlbumSortOrder(a.sortOrder.Selected)))
		a.grid.Refresh()
	}
}

func (a *AlbumsPage) Save() SavedPage {
	return &savedAlbumsPage{
		title:     a.title,
		lm:        a.lm,
		im:        a.im,
		nav:       a.nav,
		sortOrder: a.sortOrder.Selected,
		gridState: a.grid.SaveToState(),
	}
}

func (a *AlbumsPage) doSearch(query string) {
	if a.searchGrid == nil {
		a.searchGrid = widgets.NewAlbumGrid(a.lm.SearchIter(query), a.im, false /*showYear*/)
		a.searchGrid.OnPlayAlbum = a.onPlayAlbum
		a.searchGrid.OnShowAlbumPage = a.onShowAlbumPage
		a.searchGrid.OnShowArtistPage = a.onShowArtistPage
	} else {
		a.searchGrid.Reset(a.lm.SearchIter(query))
	}
	a.container.Objects[0] = a.searchGrid
	a.Refresh()
}

func (a *AlbumsPage) onPlayAlbum(albumID string) {
	if a.OnPlayAlbum != nil {
		a.OnPlayAlbum(albumID, 0)
	}
}

func (a *AlbumsPage) onShowArtistPage(artistID string) {
	a.nav(ArtistRoute(artistID))
}

func (a *AlbumsPage) onShowAlbumPage(albumID string) {
	a.nav(AlbumRoute(albumID))
}

func (a *AlbumsPage) onSortOrderChanged(order string) {
	a.grid.Reset(a.lm.AlbumsIter(backend.AlbumSortOrder(order)))
	if a.searchText == "" {
		a.container.Objects[0] = a.grid
		a.Refresh()
	}
}

func (a *AlbumsPage) CreateRenderer() fyne.WidgetRenderer {
	a.ExtendBaseWidget(a)
	return widget.NewSimpleRenderer(a.container)
}

type savedAlbumsPage struct {
	title     string
	lm        *backend.LibraryManager
	im        *backend.ImageManager
	nav       func(Route)
	sortOrder string
	gridState widgets.AlbumGridState
}

func (s *savedAlbumsPage) Restore() Page {
	return restoreAlbumsPage(s)
}