window.$ = require('jquery');
var $ = require('jquery');
var React = require('react');
var BS = require('react-bootstrap');
var Navbar = BS.Navbar;
var Nav = BS.Nav;
var NavItem = BS.NavItem;
var Grid = BS.Grid;
var Row = BS.Row;
var Colm = BS.Col;
var Button = BS.Button;
var Glyphicon = BS.Glyphicon;
/* var OverlayTrigger = BS.OverlayTrigger;
   var Popover = BS.Popover; */
var ListGroup = BS.ListGroup;
var ListGroupItem = BS.ListGroupItem;

module.exports = React.createClass({
  getDefaultProps: function() {
    return {
      navActiveKey: "home",
    };
  },

  getInitialState: function() {
    return {
      searchUsedTime: null,
      queryWeights: null,
      newss: []
    };
  },

  componentDidMount: function() {
    document.title = "newsxu - 搜尋";
  },

  handleSubmit: function(event) {
    event.preventDefault();
    var searchText = this.refs.searchInput.getDOMNode().value.trim();
    console.log(searchText);
    $.getJSON('/json', {text: searchText}, function(data) {
      console.log(data);
      this.setState({newss: data.newss, searchUsedTime: data.searchUsedTime});
    }.bind(this));
  },

  toNews: function(news) {
    return (
      <ListGroupItem key={news.title} href={news.url}>
        <h3 className="news-title"> { news.title } </h3>
        <small>{news.provider} - {news.postTime}</small>
      </ListGroupItem>
    );
  },

  renderSearchResultStats: function() {
    if (this.state.newss.length === 0) {
      return null;
    } else {
      return (
        <div style={{marginBottom: "20px"}}>
          <small> {this.state.newss.length} 項結果 (搜尋時間： {this.state.searchUsedTime}) </small>
        </div>
      );
    }
  },

  render: function() {
    var brand = (<a href="#home">newsxu</a>);
    var style = {marginLeft: "9%"};
    return (
      <div>
        <Navbar brand={brand} fluid>
          <div className="collapse navbar-collapse">
            <form className="navbar-form navbar-left"
                  style={style} role="search"
                  onSubmit={this.handleSubmit}>
              <div className="form-group">
                <input ref="searchInput" type="text"
                       style={{width: "35vw"}}
                       autoFocus className="form-control nav-news-search-bar" placeholder="Search" />
                <Button type="submit"><Glyphicon glyph="search" /></Button>
              </div>
            </form>
            <Nav navbar right activeKey={this.props.navActiveKey}>
              <NavItem eventKey="about" href="#about">
                關於
              </NavItem>
            </Nav>
          </div>
        </Navbar>
        <Grid fluid>
          <Row>
            <Colm xs={2} />
            <Colm xs={6}>
              { this.renderSearchResultStats() }
              <div className="news-list-group">
                <ListGroup>
                  { this.state.newss.map(this.toNews) }
                </ListGroup>
              </div>
            </Colm>
            <Colm xs={4} />
          </Row>
        </Grid>
      </div>
    );
  }
});
