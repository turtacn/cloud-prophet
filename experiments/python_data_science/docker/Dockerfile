FROM ubuntu:18.04

RUN apt-get -qq update && apt-get -qq -y install vim curl bzip2 build-essential \
    && curl -sSL https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh -o /tmp/miniconda.sh \
    && bash /tmp/miniconda.sh -bfp /usr/local \
    && rm -rf /tmp/miniconda.sh \
    && conda install -y python=3 \
    && conda update conda \
    && apt-get -qq -y remove curl bzip2 \
    && apt-get -qq -y autoremove \
    && apt-get autoclean \
    && rm -rf /var/lib/apt/lists/* /var/log/dpkg.log \
    && conda clean --all --yes

ENV PATH /opt/conda/bin:$PATH


RUN conda install -y \
    h5py \
    pandas \
    scikit-learn \
    matplotlib \
    seaborn \
    theano \
    jupyter \
    tensorflow \
    keras

RUN conda install -c conda-forge fbprophet
RUN conda install -c anaconda scikit-learn
RUN conda install -c anaconda dill dask
RUN python   -m   'pip'  install gsutil

# then  go to [clean]

# install opencv
##RUN conda install --channel https://conda.anaconda.org/menpo opencv3

# [clean] clean here
RUN conda clean --yes --tarballs --packages --source-cache

VOLUME /notebook
WORKDIR /notebook
EXPOSE 8888

CMD jupyter notebook --no-browser --ip=0.0.0.0 --allow-root --NotebookApp.token=