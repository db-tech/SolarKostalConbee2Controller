import React from 'react';

const ErrorPage = ({response}) => {
    return (
        <div className="container text-center mt-5">
            <h1 className="display-4 mb-4">Oops! Something went wrong</h1>
            <p className="lead mb-5">We're sorry, but it looks like an error has occurred. {response.StatusMessage}</p>
        </div>
    );
};

export default ErrorPage;
